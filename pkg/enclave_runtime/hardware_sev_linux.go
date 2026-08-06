//go:build linux

package enclave_runtime

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"

	sevabi "github.com/virtengine/virtengine/pkg/enclave_runtime/sev"
)

const sevGuestMsgVersion = 1

type linuxSNPReportRequest struct {
	UserData [SNP_REPORT_DATA_SIZE]byte
	VMPL     uint32
	Reserved [28]byte
}

type linuxSNPReportResponse struct {
	Data [4000]byte
}

type linuxSNPDerivedKeyRequest struct {
	RootKeySelect    uint32
	Reserved         uint32
	GuestFieldSelect uint64
	VMPL             uint32
	GuestSVN         uint32
	TCBVersion       uint64
}

type linuxSNPDerivedKeyResponse struct {
	Data [64]byte
}

type linuxSNPGuestRequestIOCTL struct {
	MsgVersion uint8
	_          [7]byte
	ReqData    uint64
	RespData   uint64
	ExitInfo2  uint64
}

func requestSEVHardwareReport(fd *os.File, userData [64]byte, vmpl uint32) ([]byte, error) {
	req := linuxSNPReportRequest{
		UserData: userData,
		VMPL:     vmpl,
	}
	resp := linuxSNPReportResponse{}
	if err := sevGuestIoctl(fd, SNP_GET_REPORT, unsafe.Pointer(&req), unsafe.Pointer(&resp)); err != nil {
		return nil, err
	}

	return append([]byte(nil), resp.Data[:sevabi.ReportSize]...), nil
}

func requestSEVDerivedKey(fd *os.File, rootKey int, guestFieldSelect uint64, vmpl uint32) ([]byte, error) {
	req := linuxSNPDerivedKeyRequest{
		RootKeySelect:    uint32(rootKey),
		GuestFieldSelect: guestFieldSelect,
		VMPL:             vmpl,
	}
	resp := linuxSNPDerivedKeyResponse{}
	if err := sevGuestIoctl(fd, SNP_GET_DERIVED_KEY, unsafe.Pointer(&req), unsafe.Pointer(&resp)); err != nil {
		return nil, err
	}

	return append([]byte(nil), resp.Data[:32]...), nil
}

func sevGuestIoctl(fd *os.File, request uintptr, reqPtr, respPtr unsafe.Pointer) error {
	if fd == nil {
		return ErrHardwareNotInitialized
	}
	if reqPtr == nil || respPtr == nil {
		return ErrHardwareOperationUnsupported
	}

	ioctlReq := linuxSNPGuestRequestIOCTL{
		MsgVersion: sevGuestMsgVersion,
		ReqData:    uint64(uintptr(reqPtr)),
		RespData:   uint64(uintptr(respPtr)),
	}

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd.Fd(), request, uintptr(unsafe.Pointer(&ioctlReq)))
	if errno != 0 {
		return os.NewSyscallError("ioctl", errno)
	}
	if ioctlReq.ExitInfo2 != 0 {
		fwErr := uint32(ioctlReq.ExitInfo2)
		vmmErr := uint32(ioctlReq.ExitInfo2 >> 32)
		return fmt.Errorf("snp guest request failed: fw_error=%d vmm_error=%d", fwErr, vmmErr)
	}

	return nil
}
