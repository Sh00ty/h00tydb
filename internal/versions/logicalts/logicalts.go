package logicalts

import (
	"fmt"
	"strconv"
	"strings"
)

type Version struct {
	SeqNum         uint64
	InternalSeqNum uint64
	// must be once generated uuid
	// plus node live term
	// for correct restarts
	NodeID string
}

func (v Version) String() string {
	return fmt.Sprintf("%d-%s-%d", v.SeqNum, v.NodeID, v.InternalSeqNum)
}

func (v Version) IsNull() bool {
	nullVer := Version{}
	return v == nullVer
}

func (v Version) IsBigger(ver Version) bool {
	if v.SeqNum != ver.SeqNum {
		return v.SeqNum > ver.SeqNum
	}
	switch strings.Compare(v.NodeID, ver.NodeID) {
	case 1:
		return true
	case -1:
		return false
	case 0:
	}
	return v.InternalSeqNum > ver.InternalSeqNum
}

type VersionBuilder struct {
}

func (vb VersionBuilder) FromString(str string) (Version, error) {
	seqIdx := strings.Index(str, "-")
	if seqIdx == -1 || seqIdx == len(str) {
		return Version{}, fmt.Errorf("wrong version format detected: seq: %s", str)
	}
	seqNum, err := strconv.ParseUint(str[:seqIdx], 10, 64)
	if err != nil {
		return Version{}, fmt.Errorf("can't parse seq %s: %w", str, err)
	}

	internalSeqIdx := strings.LastIndex(str, "-")
	if internalSeqIdx == -1 || internalSeqIdx == len(str) {
		return Version{}, fmt.Errorf("wrong version format detected: internalSeq: %s", str)
	}

	internalSeq, err := strconv.ParseUint(str[internalSeqIdx+1:], 10, 64)
	if err != nil {
		return Version{}, fmt.Errorf("can't parse seq %s: %w", str, err)
	}
	nodeID := str[seqIdx+1 : internalSeqIdx]
	return Version{
		SeqNum:         seqNum,
		NodeID:         nodeID,
		InternalSeqNum: internalSeq,
	}, nil
}
