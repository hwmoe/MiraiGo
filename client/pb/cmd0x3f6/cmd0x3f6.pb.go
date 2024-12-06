// Code generated by protoc-gen-golite. DO NOT EDIT.
// source: pb/cmd0x3f6/cmd0x3f6.proto

package cmd0x3f6

import (
	proto "github.com/RomiChan/protobuf/proto"
)

type C3F6ReqBody struct {
	SubCmd                            proto.Option[uint32]               `protobuf:"varint,1,opt"`
	CrmCommonHead                     *C3F6CRMMsgHead                    `protobuf:"bytes,2,opt"`
	SubcmdLoginProcessCompleteReqBody *QDUserLoginProcessCompleteReqBody `protobuf:"bytes,42,opt"`
	GetQQFriendGroupListReqBody       *GetQQFriendGroupListReqBody       `protobuf:"bytes,254,opt"`
	_                                 [0]func()
}

type C3F6RspBody struct {
	SubCmd                            proto.Option[uint32]               `protobuf:"varint,1,opt"`
	CrmCommonHead                     *C3F6CRMMsgHead                    `protobuf:"bytes,2,opt"`
	SubcmdLoginProcessCompleteRspBody *QDUserLoginProcessCompleteRspBody `protobuf:"bytes,42,opt"`
	GetQQFriendGroupListRspBody       *GetQQFriendGroupListRspBody       `protobuf:"bytes,254,opt"`
	_                                 [0]func()
}

type QDUserLoginProcessCompleteReqBody struct {
	Kfext        proto.Option[uint64] `protobuf:"varint,1,opt"`
	Pubno        proto.Option[uint32] `protobuf:"varint,2,opt"`
	Buildno      proto.Option[uint32] `protobuf:"varint,3,opt"`
	TerminalType proto.Option[uint32] `protobuf:"varint,4,opt"`
	Status       proto.Option[uint32] `protobuf:"varint,5,opt"`
	LoginTime    proto.Option[uint32] `protobuf:"varint,6,opt"`
	HardwareInfo proto.Option[string] `protobuf:"bytes,7,opt"`
	SoftwareInfo proto.Option[string] `protobuf:"bytes,8,opt"`
	Guid         []byte               `protobuf:"bytes,9,opt"`
	AppName      proto.Option[string] `protobuf:"bytes,10,opt"`
	SubAppId     proto.Option[uint32] `protobuf:"varint,11,opt"`
}

type QDUserLoginProcessCompleteRspBody struct {
	Ret                *RetInfo             `protobuf:"bytes,1,opt"`
	Url                proto.Option[string] `protobuf:"bytes,2,opt"`
	Mobile             proto.Option[string] `protobuf:"bytes,3,opt"`
	ExternalMobile     proto.Option[string] `protobuf:"bytes,4,opt"`
	DataAnalysisPriv   proto.Option[bool]   `protobuf:"varint,5,opt"`
	DeviceLock         proto.Option[bool]   `protobuf:"varint,6,opt"`
	ModulePrivilege    proto.Option[uint64] `protobuf:"varint,7,opt"`
	ModuleSubPrivilege []uint32             `protobuf:"varint,8,rep"`
	MasterSet          proto.Option[uint32] `protobuf:"varint,9,opt"`
	ExtSet             proto.Option[uint32] `protobuf:"varint,10,opt"`
	CorpConfProperty   proto.Option[uint64] `protobuf:"varint,11,opt"`
	Corpuin            proto.Option[uint64] `protobuf:"varint,12,opt"`
	Kfaccount          proto.Option[uint64] `protobuf:"varint,13,opt"`
	SecurityLevel      proto.Option[uint32] `protobuf:"varint,14,opt"`
	MsgTitle           proto.Option[string] `protobuf:"bytes,15,opt"`
	SuccNoticeMsg      proto.Option[string] `protobuf:"bytes,16,opt"`
	NameAccount        proto.Option[uint64] `protobuf:"varint,17,opt"`
	CrmMigrateFlag     proto.Option[uint32] `protobuf:"varint,18,opt"`
	ExtuinName         proto.Option[string] `protobuf:"bytes,19,opt"`
	OpenAccountTime    proto.Option[uint32] `protobuf:"varint,20,opt"`
}

type RetInfo struct {
	RetCode  proto.Option[uint32] `protobuf:"varint,1,opt"`
	ErrorMsg proto.Option[string] `protobuf:"bytes,2,opt"`
	_        [0]func()
}

type C3F6CRMMsgHead struct {
	CrmSubCmd  proto.Option[uint32] `protobuf:"varint,1,opt"`
	HeadLen    proto.Option[uint32] `protobuf:"varint,2,opt"`
	VerNo      proto.Option[uint32] `protobuf:"varint,3,opt"`
	KfUin      proto.Option[uint64] `protobuf:"varint,4,opt"`
	Seq        proto.Option[uint32] `protobuf:"varint,5,opt"`
	PackNum    proto.Option[uint32] `protobuf:"varint,6,opt"`
	CurPack    proto.Option[uint32] `protobuf:"varint,7,opt"`
	BufSig     proto.Option[string] `protobuf:"bytes,8,opt"`
	Clienttype proto.Option[uint32] `protobuf:"varint,9,opt"`
	LaborUin   proto.Option[uint64] `protobuf:"varint,10,opt"`
	LaborName  proto.Option[string] `protobuf:"bytes,11,opt"`
	Kfaccount  proto.Option[uint64] `protobuf:"varint,12,opt"`
	TraceId    proto.Option[string] `protobuf:"bytes,13,opt"`
	AppId      proto.Option[uint32] `protobuf:"varint,14,opt"`
	_          [0]func()
}

type GetQQFriendGroupListReqBody struct {
}

type QQFriendGroupInfo struct {
	GroupId    proto.Option[uint32] `protobuf:"varint,1,opt"`
	GroupName  proto.Option[string] `protobuf:"bytes,2,opt"`
	CreateTime proto.Option[uint32] `protobuf:"fixed32,3,opt"`
}

type QQFriendGroupInfoListBody struct {
	QQFriendGroupInfo []*QQFriendGroupInfo `protobuf:"bytes,1,rep"`
}

type GetQQFriendGroupListRspBody struct {
	Ret                   *RetInfo                   `protobuf:"bytes,1,opt"`
	QQFriendGroupInfoList *QQFriendGroupInfoListBody `protobuf:"bytes,2,opt"`
}
