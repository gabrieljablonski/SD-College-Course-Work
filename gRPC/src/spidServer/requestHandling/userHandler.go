package requestHandling

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"log"
	"spidServer/entities"
	"spidServer/gps"
	pb "spidServer/requestHandling/protoBuffers"
)

func (h *Handler) GetUserInfo(ctx context.Context, request *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	user, err :=  h.queryUser(request.UserID)
	if err != nil {
		err = fmt.Errorf("failed to get user info: %s", err)
		log.Print(err)
		return nil, err
	}
	return &pb.GetUserResponse{
		Message: "User queried successfully.",
		User:    user,
	}, nil
}

func (h *Handler) RegisterUser(ctx context.Context, request *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error) {
	user, err :=  h.registerUser(request.Name, gps.FromProtoBufferEntity(request.Position))
	if err != nil {
		err = fmt.Errorf("failed to register user: %s", err)
		log.Print(err)
		return nil, err
	}
	return &pb.RegisterUserResponse{
		Message: "User registered successfully.",
		User:    user,
	}, nil
}

func (h *Handler) UpdateUser(ctx context.Context, request *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	user, err :=  h.queryUser(request.User.Id)
	if err != nil {
		err = fmt.Errorf("failed to update user: %s", err)
		log.Print(err)
		return nil, err
	}
	err = h.updateUser(user)
	if err != nil {
		return nil, fmt.Errorf("failed to update user position: %s", err)
	}
	return &pb.UpdateUserResponse{
		Message: "User position updated successfully.",
		User:    user,
	}, nil
}

func (h *Handler) DeleteUser(ctx context.Context, request *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	user, err := h.deleteUser(request.UserID)
	if err != nil {
		err = fmt.Errorf("failed to delete user: %s", err)
		log.Print(err)
		return nil, err
	}
	return &pb.DeleteUserResponse{
		Message: "Deleted user successfully.",
		User: user,
	}, nil
}

func (h *Handler) RequestAssociation(ctx context.Context, request *pb.RequestAssociationRequest) (*pb.RequestAssociationResponse, error) {
	errPrefix := "failed to request association"
	pbUser, err := h.queryUser(request.UserID)
	if err != nil {
		err = fmt.Errorf("%s: %s", errPrefix, err)
		log.Print(err)
		return nil, err
	}
	user, err := entities.UserFromProtoBufferEntity(pbUser)
	if err != nil {
		err = fmt.Errorf("%s: %s", errPrefix, err)
		log.Print(err)
		return nil, err
	}
	if user.CurrentSpidID != uuid.Nil {
		err = fmt.Errorf("%s: user is already associated to spid with id `%s`", errPrefix, user.CurrentSpidID)
		log.Print(err)
		return nil, err
	}
	pbSpid, err := h.querySpid(request.SpidID)
	if err != nil {
		err = fmt.Errorf("%s: %s", errPrefix, err)
		log.Print(err)
		return nil, err
	}
	spid, err := entities.SpidFromProtoBufferEntity(pbSpid)
	if err != nil {
		err = fmt.Errorf("%s: %s", errPrefix, err)
		log.Print(err)
		return nil, err
	}
	if spid.CurrentUserID != uuid.Nil {
		err = fmt.Errorf("%s: spid is already associated to user with id `%s`", errPrefix, spid.CurrentUserID)
		log.Print(err)
		return nil, err
	}
	user.CurrentSpidID = spid.ID
	spid.CurrentUserID = user.ID
	err = h.updateUser(user.ToProtoBufferEntity())
	if err != nil {
		err = fmt.Errorf("%s: %s", errPrefix, err)
		log.Print(err)
		return nil, err
	}
	err = h.updateSpid(spid.ToProtoBufferEntity())
	if err != nil {
		// if update spid failed, rollback update user
		user.CurrentSpidID = uuid.Nil
		err2 := h.updateUser(user.ToProtoBufferEntity())
		if err2 != nil {
			// if this ever happens, server will hold inconsistent data
			err = fmt.Errorf("%s: failed to rollback `%s`, `%s`", errPrefix, err, err2)
			log.Print(err)
			return nil, err
		}
		err = fmt.Errorf("%s: %s", errPrefix, err)
		log.Print(err)
		return nil, err
	}
	return &pb.RequestAssociationResponse{
		Message: "Association request successful.",
		User:    user.ToProtoBufferEntity(),
	}, nil
}

func (h *Handler) RequestDissociation(ctx context.Context, request *pb.RequestDissociationRequest) (*pb.RequestDissociationResponse, error) {
	errPrefix := "failed to request association"
	pbUser, err := h.queryUser(request.UserID)
	if err != nil {
		err = fmt.Errorf("%s: %s", errPrefix, err)
		log.Print(err)
		return nil, err
	}
	user, err := entities.UserFromProtoBufferEntity(pbUser)
	if err != nil {
		err = fmt.Errorf("%s: %s", errPrefix, err)
		log.Print(err)
		return nil, err
	}
	if user.CurrentSpidID == uuid.Nil {
		err = fmt.Errorf("%s: user is not associated to any spids", errPrefix)
		log.Print(err)
		return nil, err
	}
	pbSpid, err := h.querySpid(user.CurrentSpidID.String())
	if err != nil {
		err = fmt.Errorf("%s: %s", errPrefix, err)
		log.Print(err)
		return nil, err
	}
	spid, err := entities.SpidFromProtoBufferEntity(pbSpid)
	if err != nil {
		err = fmt.Errorf("%s: %s", errPrefix, err)
		log.Print(err)
		return nil, err
	}
	if spid.CurrentUserID != user.ID {
		err = fmt.Errorf("%s: spid is not associated to user", errPrefix)
		log.Print(err)
		return nil, err
	}
	user.CurrentSpidID = uuid.Nil
	spid.CurrentUserID = uuid.Nil
	err = h.updateUser(user.ToProtoBufferEntity())
	if err != nil {
		err = fmt.Errorf("%s: %s", errPrefix, err)
		log.Print(err)
		return nil, err
	}
	err = h.updateSpid(spid.ToProtoBufferEntity())
	if err != nil {
		// if update spid failed, rollback update user
		user.CurrentSpidID = uuid.Nil
		err2 := h.updateUser(user.ToProtoBufferEntity())
		if err2 != nil {
			// if this ever happens, server will hold inconsistent data
			err = fmt.Errorf("%s: failed to rollback `%s`, `%s`", errPrefix, err, err2)
			log.Print(err)
			return nil, err
		}
		err = fmt.Errorf("%s: %s", errPrefix, err)
		log.Print(err)
		return nil, err
	}
	return &pb.RequestDissociationResponse{
		Message: "Dissociation request successful.",
		User:    user.ToProtoBufferEntity(),
	}, nil
}

func (h *Handler) RequestSpidInfo(ctx context.Context, request *pb.RequestSpidInfoRequest) (*pb.RequestSpidInfoResponse, error) {
	errPrefix := "failed to request spid info"
	pbUser, err := h.queryUser(request.UserID)
	if err != nil {
		err = fmt.Errorf("%s: %s", errPrefix, err)
		log.Print(err)
		return nil, err
	}
	user, err := entities.UserFromProtoBufferEntity(pbUser)
	if err != nil {
		err = fmt.Errorf("%s: %s", errPrefix, err)
		log.Print(err)
		return nil, err
	}
	pbSpid, err := h.querySpid(request.SpidID)
	if err != nil {
		err = fmt.Errorf("%s: %s", errPrefix, err)
		log.Print(err)
		return nil, err
	}
	spid, err := entities.SpidFromProtoBufferEntity(pbSpid)
	if err != nil {
		err = fmt.Errorf("%s: %s", errPrefix, err)
		log.Print(err)
		return nil, err
	}
	if user.CurrentSpidID != spid.ID || spid.CurrentUserID != user.ID {
		// if this happens, it means server data is inconsistent
		err = fmt.Errorf("%s: user with id `%s` not associated to spid with id `%s`", errPrefix, user.ID, spid.ID)
		log.Print(err)
		return nil, err
	}
	return &pb.RequestSpidInfoResponse{
		Message: "Spid info request successful.",
		Spid:    spid.ToProtoBufferEntity(),
	}, nil
}

func (h *Handler) RequestLockChange(ctx context.Context, request *pb.RequestLockChangeRequest) (*pb.RequestLockChangeResponse, error) {
	errPrefix := "failed to request lock change"
	pbUser, err := h.queryUser(request.UserID)
	if err != nil {
		err = fmt.Errorf("%s: %s", errPrefix, err)
		log.Print(err)
		return nil, err
	}
	user, err := entities.UserFromProtoBufferEntity(pbUser)
	if err != nil {
		err = fmt.Errorf("%s: %s", errPrefix, err)
		log.Print(err)
		return nil, err
	}
	pbSpid, err := h.querySpid(request.SpidID)
	if err != nil {
		err = fmt.Errorf("%s: %s", errPrefix, err)
		log.Print(err)
		return nil, err
	}
	spid, err := entities.SpidFromProtoBufferEntity(pbSpid)
	if err != nil {
		err = fmt.Errorf("%s: %s", errPrefix, err)
		log.Print(err)
		return nil, err
	}
	if user.CurrentSpidID != spid.ID || spid.CurrentUserID != user.ID {
		// if this happens, it means server data is inconsistent
		err = fmt.Errorf("%s: user with id `%s` not associated to spid with id `%s`", errPrefix, user.ID, spid.ID)
		log.Print(err)
		return nil, err
	}
	err = spid.UpdateLockState(request.LockState)
	if err != nil {
		err = fmt.Errorf("%s: %s", errPrefix, err)
		log.Print(err)
		return nil, err
	}
	err = h.updateSpid(spid.ToProtoBufferEntity())
	if err != nil {
		err = fmt.Errorf("%s: %s", errPrefix, err)
		log.Print(err)
		return nil, err
	}
	return &pb.RequestLockChangeResponse{
		Message: "Lock change request successful.",
		Spid:    spid.ToProtoBufferEntity(),
	}, nil
}

func (h *Handler) AddRemoteUser(ctx context.Context, request *pb.AddRemoteUserRequest) (*pb.AddRemoteUserResponse, error) {
	err := h.addRemoteUser(request.User)
	if err != nil {
		err = fmt.Errorf("failed to add remote user: %s", err)
		log.Print(err)
		return nil, err
	}
	return &pb.AddRemoteUserResponse{
		Message: "User added remotely successfully.",
	}, nil
}

func (h *Handler) UpdateRemoteUser(ctx context.Context, request *pb.UpdateRemoteUserRequest) (*pb.UpdateRemoteUserResponse, error) {
	err := h.updateRemoteUser(request.User)
	if err != nil {
		err = fmt.Errorf("failed to update remote user: %s", err)
		log.Print(err)
		return nil, err
	}
	return &pb.UpdateRemoteUserResponse{
		Message: "User updated remotely successfully.",
	}, nil
}

func (h *Handler) RemoveRemoteUser(ctx context.Context, request *pb.RemoveRemoteUserRequest) (*pb.RemoveRemoteUserResponse, error) {
	err := h.removeRemoteUser(request.UserID)
	if err != nil {
		err = fmt.Errorf("failed to remove remote user: %s", err)
		log.Print(err)
		return nil, err
	}
	return &pb.RemoveRemoteUserResponse{
		Message: "User removed remotely successfully.",
	}, nil
}
