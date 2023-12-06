// Code generated by protoc-gen-golite. DO NOT EDIT.
// source: pb/cmd0x6ff/smbcmd0x519.proto

package cmd0x6ff

import (
	proto "github.com/RomiChan/protobuf/proto"
)

type C519CRMMsgHead struct {
	CrmSubCmd  proto.Option[uint32] `protobuf:"varint,1,opt"`
	HeadLen    proto.Option[uint32] `protobuf:"varint,2,opt"`
	VerNo      proto.Option[uint32] `protobuf:"varint,3,opt"`
	KfUin      proto.Option[uint64] `protobuf:"varint,4,opt"`
	Seq        proto.Option[uint32] `protobuf:"varint,5,opt"`
	PackNum    proto.Option[uint32] `protobuf:"varint,6,opt"`
	CurPack    proto.Option[uint32] `protobuf:"varint,7,opt"`
	BufSig     proto.Option[string] `protobuf:"bytes,8,opt"`
	PubQq      proto.Option[uint64] `protobuf:"varint,9,opt"`
	Clienttype proto.Option[uint32] `protobuf:"varint,10,opt"`
	LaborUin   proto.Option[uint64] `protobuf:"varint,11,opt"`
	LaborName  proto.Option[string] `protobuf:"bytes,12,opt"`
	Puin       proto.Option[uint64] `protobuf:"varint,13,opt"`
	LoginSig   *LoginSig            `protobuf:"bytes,18,rep"`
	_          [0]func()
}

type LoginSig struct {
	Type proto.Option[uint32] `protobuf:"varint,1,opt"`
	Sig  []byte               `protobuf:"bytes,2,opt"`
}

type GetNavigationMenuReqBody struct {
	Puin  proto.Option[uint64] `protobuf:"varint,1,opt"`
	Uin   proto.Option[uint64] `protobuf:"varint,2,opt"`
	VerNo proto.Option[uint32] `protobuf:"varint,3,opt"`
	_     [0]func()
}

type GetNavigationMenuRspBody struct {
	Ret    *C519RetInfo         `protobuf:"bytes,1,opt"`
	IsShow proto.Option[int32]  `protobuf:"varint,2,opt"`
	UctMsg proto.Option[string] `protobuf:"bytes,3,opt"`
	VerNo  proto.Option[uint32] `protobuf:"varint,4,opt"`
	_      [0]func()
}

type C519ReqBody struct {
	SubCmd                      proto.Option[uint32]         `protobuf:"varint,1,opt"`
	CrmCommonHead               *C519CRMMsgHead              `protobuf:"bytes,2,opt"`
	GetAddressDetailListReqBody *GetAddressDetailListReqBody `protobuf:"bytes,33,opt"` // optional GetNavigationMenuReqBody getNavigationMenuReq = 35;
	_                           [0]func()
}

type C519RetInfo struct {
	RetCode  proto.Option[uint32] `protobuf:"varint,1,opt"`
	ErrorMsg proto.Option[string] `protobuf:"bytes,2,opt"`
	_        [0]func()
}

type C519RspBody struct {
	SubCmd proto.Option[uint32] `protobuf:"varint,1,opt"`
	// optional C519CRMMsgHead crmCommonHead = 2;
	GetAddressDetailListRspBody *GetAddressDetailListRspBody `protobuf:"bytes,33,opt"` //optional GetNavigationMenuRspBody getNavigationMenuRsp = 35;
	_                           [0]func()
}

type GetAddressDetailListReqBody struct {
	Timestamp  proto.Option[uint32] `protobuf:"fixed32,1,opt"`
	Timestamp2 proto.Option[uint64] `protobuf:"fixed64,2,opt"`
	_          [0]func()
}

type GetAddressDetailListRspBody struct {
	// optional C519RetInfo ret = 1;
	Timestamp     proto.Option[uint32] `protobuf:"fixed32,2,opt"`
	Full          proto.Option[bool]   `protobuf:"varint,3,opt"`
	AddressDetail []*AddressDetail     `protobuf:"bytes,4,rep"`
	Timestamp2    proto.Option[uint64] `protobuf:"fixed64,5,opt"`
}

type AddressDetail struct {
	Aid              proto.Option[uint32] `protobuf:"varint,1,opt"`
	ModifyTime       proto.Option[uint32] `protobuf:"fixed32,2,opt"`
	CreateTime       proto.Option[uint32] `protobuf:"fixed32,3,opt"`
	Status           proto.Option[uint32] `protobuf:"varint,4,opt"`
	Groupid          proto.Option[uint32] `protobuf:"varint,5,opt"`
	AddGroupName     []byte               `protobuf:"bytes,6,opt"`
	Name             []byte               `protobuf:"bytes,7,opt"`
	Gender           proto.Option[uint32] `protobuf:"varint,8,opt"`
	Birthday         proto.Option[uint32] `protobuf:"fixed32,9,opt"`
	Company0         []byte               `protobuf:"bytes,10,opt"`
	CompanyPosition0 []byte               `protobuf:"bytes,11,opt"`
	Company1         []byte               `protobuf:"bytes,12,opt"`
	CompanyPosition1 []byte               `protobuf:"bytes,13,opt"`
	FixedPhone0      []byte               `protobuf:"bytes,14,opt"`
	FixedPhone1      []byte               `protobuf:"bytes,15,opt"`
	Email0           []byte               `protobuf:"bytes,16,opt"`
	Email1           []byte               `protobuf:"bytes,17,opt"`
	Fax0             []byte               `protobuf:"bytes,18,opt"`
	Fax1             []byte               `protobuf:"bytes,19,opt"`
	Comment          []byte               `protobuf:"bytes,20,opt"`
	HeadUrl          []byte               `protobuf:"bytes,21,opt"`
	// repeated AddressMobileInfo mobilePhone = 22;
	MobilePhoneUpdated proto.Option[bool]   `protobuf:"varint,23,opt"`
	Qq                 []*AddressQQinfo     `protobuf:"bytes,24,rep"`
	QqPhoneUpdated     proto.Option[bool]   `protobuf:"varint,25,opt"`
	ModifyTime2        proto.Option[uint64] `protobuf:"fixed64,26,opt"`
}

type AddressMobileInfo struct {
	Index            proto.Option[uint32] `protobuf:"varint,1,opt"`
	Account          []byte               `protobuf:"bytes,2,opt"`
	FormattedAccount []byte               `protobuf:"bytes,5,opt"`
}

type AddressQQinfo struct {
	Index   proto.Option[uint32] `protobuf:"varint,1,opt"`
	Account proto.Option[uint64] `protobuf:"varint,2,opt"`
	_       [0]func()
}

type NewBizClientRegion struct {
	ClientNation   proto.Option[string] `protobuf:"bytes,1,opt"`
	ClientProvince proto.Option[string] `protobuf:"bytes,2,opt"`
	ClientCity     proto.Option[string] `protobuf:"bytes,3,opt"`
	ClientRegion   proto.Option[string] `protobuf:"bytes,4,opt"`
	_              [0]func()
}

type NewBizClientRegionCode struct {
	Nationid   proto.Option[uint64] `protobuf:"varint,1,opt"`
	Provinceid proto.Option[uint64] `protobuf:"varint,2,opt"`
	Cityid     proto.Option[uint64] `protobuf:"varint,3,opt"`
	Regionid   proto.Option[uint64] `protobuf:"varint,4,opt"`
	_          [0]func()
}
