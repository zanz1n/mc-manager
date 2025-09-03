package db

import (
	"github.com/zanz1n/mc-manager/internal/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (u *User) IntoPB() *pb.User {
	return &pb.User{
		Id:            uint64(u.ID),
		CreatedAt:     timestamppb.New(u.CreatedAt),
		UpdatedAt:     timestamppb.New(u.UpdatedAt),
		Username:      u.Username,
		FirstName:     u.FirstName,
		LastName:      u.LastName,
		MinecraftUser: u.MinecraftUser,
		Email:         u.Email,
		Admin:         u.Admin,
		TwoFa:         u.TwoFa,
	}
}

func (n *Node) IntoPB() *pb.Node {
	return &pb.Node{
		Id:          uint64(n.ID),
		CreatedAt:   timestamppb.New(n.CreatedAt),
		UpdatedAt:   timestamppb.New(n.UpdatedAt),
		Name:        n.Name,
		Description: n.Description,
		Maintenance: n.Maintenance,
		Token:       n.Token,
		Endpoint:    n.Endpoint,
		EndpointTls: n.EndpointTls,
		FtpPort:     n.FtpPort,
		GrpcPort:    n.GrpcPort,
	}
}

func (i *Instance) IntoPB(state pb.InstanceState, players int32) *pb.Instance {
	ll := (*timestamppb.Timestamp)(nil)
	if i.LastLaunched.Valid {
		ll = timestamppb.New(i.LastLaunched.Time)
	}

	return &pb.Instance{
		Id:            uint64(i.ID),
		UserId:        uint64(i.UserID),
		NodeId:        uint64(i.NodeID),
		CreatedAt:     timestamppb.New(i.CreatedAt),
		UpdatedAt:     timestamppb.New(i.UpdatedAt),
		LastLaunched:  ll,
		State:         state,
		Players:       players,
		Name:          i.Name,
		Description:   i.Description,
		Version:       i.Version,
		VersionDistro: i.VersionDistro,
		Maintenance:   i.Maintenance,
		Config:        i.Config,
		Limits:        i.Limits,
	}
}
