package requestHandling

import (
	"context"
	"fmt"
	"spidServer/gps"
	pb "spidServer/requestHandling/protoBuffers"
)

func (h *Handler) GetSpidInfo(ctx context.Context, request *pb.GetSpidRequest) (*pb.GetSpidResponse, error) {
	spid, err :=  h.querySpid(request.SpidID)
	if err != nil {
		return nil, fmt.Errorf("failed to get spid info: %s", err)
	}
	return &pb.GetSpidResponse{
		Message: "Spid queried successfully.",
		Spid:    spid,
	}, nil
}

func (h *Handler) RegisterSpid(ctx context.Context, request *pb.RegisterSpidRequest) (*pb.RegisterSpidResponse, error) {
	spid, err :=  h.registerSpid(request.BatteryLevel, gps.FromProtoBufferEntity(request.Location))
	if err != nil {
		return nil, fmt.Errorf("failed to register spid")
	}
	return &pb.RegisterSpidResponse{
		Message: "Spid registered successfully",
		Spid:    spid,
	}, nil
}

func (h *Handler) UpdateSpid(ctx context.Context, request *pb.UpdateSpidRequest) (*pb.UpdateSpidResponse, error) {
	spid, err := h.querySpid(request.Spid.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to update spid location: %s", err)
	}
	err = h.updateSpid(spid)
	if err != nil {
		return nil, fmt.Errorf("failed to update spid location: %s", err)
	}
	return &pb.UpdateSpidResponse{
		Message: "Spid location updated successfully.",
		Spid:    spid,
	}, nil
}

func (h *Handler) DeleteSpid(ctx context.Context, request *pb.DeleteSpidRequest) (*pb.DeleteSpidResponse, error) {
	spid, err := h.querySpid(request.SpidID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete spid: %s", err)
	}
	err = h.deleteSpid(spid)
	if err != nil {
		return nil, fmt.Errorf("failed to delete spid: %s", err)
	}
	return &pb.DeleteSpidResponse{
		Message: "Spid deleted successfully.",
		Spid:    spid,
	}, nil
}
