package utils

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/zanz1n/mc-manager/config"
)

func Serve(ctx context.Context, cfg config.ServerConfig, handler http.Handler) error {
	start := time.Now()

	var (
		err       error
		tlsConfig *tls.Config = nil
	)
	if cfg.TLS != nil && cfg.TLS.Enable {
		tlsConfig = &tls.Config{
			Certificates: make([]tls.Certificate, 1),
		}
		tlsConfig.Certificates[0], err = getTLS(cfg.TLS)
		if err != nil {
			return fmt.Errorf("failed to load TLS: %w", err)
		}
	}

	addr := &net.TCPAddr{
		IP:   cfg.IP,
		Port: int(cfg.Port),
	}

	ln, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed listen at `%s`: %w", addr.String(), err)
	}
	defer ln.Close()

	slog.Info(
		"GRPC: Listening",
		"addr", addr,
		"took", time.Since(start).Round(time.Microsecond),
	)

	protocols := new(http.Protocols)
	protocols.SetHTTP1(true)
	if cfg.TLS != nil && cfg.TLS.Enable {
		protocols.SetHTTP2(true)
	} else {
		protocols.SetUnencryptedHTTP2(true)
	}

	s := &http.Server{
		Addr:      addr.String(),
		Handler:   handler,
		Protocols: protocols,
		TLSConfig: tlsConfig,
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
	}

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()

		shutdownStart := time.Now()

		stopctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		err := s.Shutdown(stopctx)
		slog.Info(
			"GRPC: Closed server",
			"took", time.Since(shutdownStart).Round(time.Millisecond),
			"online_for", time.Since(start).Round(time.Second),
			"error", err,
		)
	}()

	if cfg.TLS != nil && cfg.TLS.Enable {
		err = s.ServeTLS(ln, "", "")
	} else {
		err = s.Serve(ln)
	}

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to serve http: %w", err)
	}

	wg.Wait()
	return nil
}

func getTLS(cfg *config.ServerTLSConfig) (tls.Certificate, error) {
	certBuf, err := os.ReadFile(cfg.Certificate)
	if err == nil {
		keyBuf, err := os.ReadFile(cfg.Key)
		if err != nil {
			err = fmt.Errorf("load private key at `%s`: %w", cfg.Key, err)
			return tls.Certificate{}, err
		}

		cert, err := tls.X509KeyPair(certBuf, keyBuf)
		if err != nil {
			err = fmt.Errorf("load key pair: %w", err)
		}
		return cert, err
	}

	if !cfg.AutoGenerate {
		err = fmt.Errorf("load certificate at `%s`: %w", cfg.Certificate, err)
		return tls.Certificate{}, err
	}

	if err = genTLS(cfg); err != nil {
		err = fmt.Errorf("generate key pair: %w", err)
		return tls.Certificate{}, err
	}

	cert, err := tls.LoadX509KeyPair(cfg.Certificate, cfg.Key)
	if err != nil {
		err = fmt.Errorf("load generated key pair: %w", err)
	}

	return cert, err
}

func genTLS(cfg *config.ServerTLSConfig) error {
	const expiration = 365 * 24 * time.Hour

	now := time.Now()

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return fmt.Errorf("serial number: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Acme Co"},
		},
		NotBefore:             now,
		NotAfter:              now.Add(expiration),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}

	derBytes, err := x509.CreateCertificate(
		rand.Reader,
		&template,
		&template,
		&key.PublicKey,
		key,
	)
	if err != nil {
		return err
	}

	certOut, err := os.Create(cfg.Certificate)
	if err != nil {
		return err
	}
	defer certOut.Close()

	err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if err != nil {
		return err
	}

	keyOut, err := os.OpenFile(cfg.Key, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer keyOut.Close()

	privBytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return err
	}

	err = pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	if err != nil {
		return err
	}

	return nil
}
