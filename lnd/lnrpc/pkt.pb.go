// Code generated by protoc-gen-go. DO NOT EDIT.
// source: pkt.proto

package lnrpc

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type Config struct {
	UserAgentName        string   `protobuf:"bytes,3,opt,name=user_agent_name,json=userAgentName,proto3" json:"user_agent_name,omitempty"`
	UserAgentVersion     string   `protobuf:"bytes,4,opt,name=user_agent_version,json=userAgentVersion,proto3" json:"user_agent_version,omitempty"`
	UserAgentComments    string   `protobuf:"bytes,5,opt,name=user_agent_comments,json=userAgentComments,proto3" json:"user_agent_comments,omitempty"`
	Services             string   `protobuf:"bytes,7,opt,name=services,proto3" json:"services,omitempty"`
	ProtocolVersion      uint32   `protobuf:"varint,8,opt,name=protocol_version,json=protocolVersion,proto3" json:"protocol_version,omitempty"`
	DisableRelayTx       bool     `protobuf:"varint,9,opt,name=disable_relay_tx,json=disableRelayTx,proto3" json:"disable_relay_tx,omitempty"`
	TrickleInterval      int64    `protobuf:"varint,11,opt,name=trickle_interval,json=trickleInterval,proto3" json:"trickle_interval,omitempty"`
	AllowSelfConns       bool     `protobuf:"varint,12,opt,name=allow_self_conns,json=allowSelfConns,proto3" json:"allow_self_conns,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Config) Reset()         { *m = Config{} }
func (m *Config) String() string { return proto.CompactTextString(m) }
func (*Config) ProtoMessage()    {}
func (*Config) Descriptor() ([]byte, []int) {
	return fileDescriptor_3c5f63a845a51abc, []int{0}
}

func (m *Config) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Config.Unmarshal(m, b)
}
func (m *Config) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Config.Marshal(b, m, deterministic)
}
func (m *Config) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Config.Merge(m, src)
}
func (m *Config) XXX_Size() int {
	return xxx_messageInfo_Config.Size(m)
}
func (m *Config) XXX_DiscardUnknown() {
	xxx_messageInfo_Config.DiscardUnknown(m)
}

var xxx_messageInfo_Config proto.InternalMessageInfo

func (m *Config) GetUserAgentName() string {
	if m != nil {
		return m.UserAgentName
	}
	return ""
}

func (m *Config) GetUserAgentVersion() string {
	if m != nil {
		return m.UserAgentVersion
	}
	return ""
}

func (m *Config) GetUserAgentComments() string {
	if m != nil {
		return m.UserAgentComments
	}
	return ""
}

func (m *Config) GetServices() string {
	if m != nil {
		return m.Services
	}
	return ""
}

func (m *Config) GetProtocolVersion() uint32 {
	if m != nil {
		return m.ProtocolVersion
	}
	return 0
}

func (m *Config) GetDisableRelayTx() bool {
	if m != nil {
		return m.DisableRelayTx
	}
	return false
}

func (m *Config) GetTrickleInterval() int64 {
	if m != nil {
		return m.TrickleInterval
	}
	return 0
}

func (m *Config) GetAllowSelfConns() bool {
	if m != nil {
		return m.AllowSelfConns
	}
	return false
}

type PeerDesc struct {
	BytesReceived        uint64   `protobuf:"varint,1,opt,name=bytes_received,json=bytesReceived,proto3" json:"bytes_received,omitempty"`
	BytesSent            uint64   `protobuf:"varint,2,opt,name=bytes_sent,json=bytesSent,proto3" json:"bytes_sent,omitempty"`
	LastRecv             string   `protobuf:"bytes,3,opt,name=last_recv,json=lastRecv,proto3" json:"last_recv,omitempty"`
	LastSend             string   `protobuf:"bytes,4,opt,name=last_send,json=lastSend,proto3" json:"last_send,omitempty"`
	Connected            bool     `protobuf:"varint,5,opt,name=connected,proto3" json:"connected,omitempty"`
	Addr                 string   `protobuf:"bytes,6,opt,name=addr,proto3" json:"addr,omitempty"`
	Cfg                  *Config  `protobuf:"bytes,7,opt,name=cfg,proto3" json:"cfg,omitempty"`
	Inbound              bool     `protobuf:"varint,8,opt,name=inbound,proto3" json:"inbound,omitempty"`
	Na                   string   `protobuf:"bytes,9,opt,name=na,proto3" json:"na,omitempty"`
	Id                   int32    `protobuf:"varint,10,opt,name=id,proto3" json:"id,omitempty"`
	UserAgent            string   `protobuf:"bytes,11,opt,name=user_agent,json=userAgent,proto3" json:"user_agent,omitempty"`
	Services             string   `protobuf:"bytes,12,opt,name=services,proto3" json:"services,omitempty"`
	VersionKnown         bool     `protobuf:"varint,13,opt,name=version_known,json=versionKnown,proto3" json:"version_known,omitempty"`
	AdvertisedProtoVer   uint32   `protobuf:"varint,14,opt,name=advertised_proto_ver,json=advertisedProtoVer,proto3" json:"advertised_proto_ver,omitempty"`
	ProtocolVersion      uint32   `protobuf:"varint,15,opt,name=protocol_version,json=protocolVersion,proto3" json:"protocol_version,omitempty"`
	SendHeadersPreferred bool     `protobuf:"varint,16,opt,name=send_headers_preferred,json=sendHeadersPreferred,proto3" json:"send_headers_preferred,omitempty"`
	VerAckReceived       bool     `protobuf:"varint,17,opt,name=ver_ack_received,json=verAckReceived,proto3" json:"ver_ack_received,omitempty"`
	WitnessEnabled       bool     `protobuf:"varint,18,opt,name=witness_enabled,json=witnessEnabled,proto3" json:"witness_enabled,omitempty"`
	WireEncoding         string   `protobuf:"bytes,19,opt,name=wire_encoding,json=wireEncoding,proto3" json:"wire_encoding,omitempty"`
	TimeOffset           int64    `protobuf:"varint,20,opt,name=time_offset,json=timeOffset,proto3" json:"time_offset,omitempty"`
	TimeConnected        string   `protobuf:"bytes,21,opt,name=time_connected,json=timeConnected,proto3" json:"time_connected,omitempty"`
	StartingHeight       int32    `protobuf:"varint,22,opt,name=starting_height,json=startingHeight,proto3" json:"starting_height,omitempty"`
	LastBlock            int32    `protobuf:"varint,23,opt,name=last_block,json=lastBlock,proto3" json:"last_block,omitempty"`
	LastAnnouncedBlock   []byte   `protobuf:"bytes,24,opt,name=last_announced_block,json=lastAnnouncedBlock,proto3" json:"last_announced_block,omitempty"`
	LastPingNonce        uint64   `protobuf:"varint,25,opt,name=last_ping_nonce,json=lastPingNonce,proto3" json:"last_ping_nonce,omitempty"`
	LastPingTime         string   `protobuf:"bytes,26,opt,name=last_ping_time,json=lastPingTime,proto3" json:"last_ping_time,omitempty"`
	LastPingMicros       int64    `protobuf:"varint,27,opt,name=last_ping_micros,json=lastPingMicros,proto3" json:"last_ping_micros,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PeerDesc) Reset()         { *m = PeerDesc{} }
func (m *PeerDesc) String() string { return proto.CompactTextString(m) }
func (*PeerDesc) ProtoMessage()    {}
func (*PeerDesc) Descriptor() ([]byte, []int) {
	return fileDescriptor_3c5f63a845a51abc, []int{1}
}

func (m *PeerDesc) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PeerDesc.Unmarshal(m, b)
}
func (m *PeerDesc) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PeerDesc.Marshal(b, m, deterministic)
}
func (m *PeerDesc) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PeerDesc.Merge(m, src)
}
func (m *PeerDesc) XXX_Size() int {
	return xxx_messageInfo_PeerDesc.Size(m)
}
func (m *PeerDesc) XXX_DiscardUnknown() {
	xxx_messageInfo_PeerDesc.DiscardUnknown(m)
}

var xxx_messageInfo_PeerDesc proto.InternalMessageInfo

func (m *PeerDesc) GetBytesReceived() uint64 {
	if m != nil {
		return m.BytesReceived
	}
	return 0
}

func (m *PeerDesc) GetBytesSent() uint64 {
	if m != nil {
		return m.BytesSent
	}
	return 0
}

func (m *PeerDesc) GetLastRecv() string {
	if m != nil {
		return m.LastRecv
	}
	return ""
}

func (m *PeerDesc) GetLastSend() string {
	if m != nil {
		return m.LastSend
	}
	return ""
}

func (m *PeerDesc) GetConnected() bool {
	if m != nil {
		return m.Connected
	}
	return false
}

func (m *PeerDesc) GetAddr() string {
	if m != nil {
		return m.Addr
	}
	return ""
}

func (m *PeerDesc) GetCfg() *Config {
	if m != nil {
		return m.Cfg
	}
	return nil
}

func (m *PeerDesc) GetInbound() bool {
	if m != nil {
		return m.Inbound
	}
	return false
}

func (m *PeerDesc) GetNa() string {
	if m != nil {
		return m.Na
	}
	return ""
}

func (m *PeerDesc) GetId() int32 {
	if m != nil {
		return m.Id
	}
	return 0
}

func (m *PeerDesc) GetUserAgent() string {
	if m != nil {
		return m.UserAgent
	}
	return ""
}

func (m *PeerDesc) GetServices() string {
	if m != nil {
		return m.Services
	}
	return ""
}

func (m *PeerDesc) GetVersionKnown() bool {
	if m != nil {
		return m.VersionKnown
	}
	return false
}

func (m *PeerDesc) GetAdvertisedProtoVer() uint32 {
	if m != nil {
		return m.AdvertisedProtoVer
	}
	return 0
}

func (m *PeerDesc) GetProtocolVersion() uint32 {
	if m != nil {
		return m.ProtocolVersion
	}
	return 0
}

func (m *PeerDesc) GetSendHeadersPreferred() bool {
	if m != nil {
		return m.SendHeadersPreferred
	}
	return false
}

func (m *PeerDesc) GetVerAckReceived() bool {
	if m != nil {
		return m.VerAckReceived
	}
	return false
}

func (m *PeerDesc) GetWitnessEnabled() bool {
	if m != nil {
		return m.WitnessEnabled
	}
	return false
}

func (m *PeerDesc) GetWireEncoding() string {
	if m != nil {
		return m.WireEncoding
	}
	return ""
}

func (m *PeerDesc) GetTimeOffset() int64 {
	if m != nil {
		return m.TimeOffset
	}
	return 0
}

func (m *PeerDesc) GetTimeConnected() string {
	if m != nil {
		return m.TimeConnected
	}
	return ""
}

func (m *PeerDesc) GetStartingHeight() int32 {
	if m != nil {
		return m.StartingHeight
	}
	return 0
}

func (m *PeerDesc) GetLastBlock() int32 {
	if m != nil {
		return m.LastBlock
	}
	return 0
}

func (m *PeerDesc) GetLastAnnouncedBlock() []byte {
	if m != nil {
		return m.LastAnnouncedBlock
	}
	return nil
}

func (m *PeerDesc) GetLastPingNonce() uint64 {
	if m != nil {
		return m.LastPingNonce
	}
	return 0
}

func (m *PeerDesc) GetLastPingTime() string {
	if m != nil {
		return m.LastPingTime
	}
	return ""
}

func (m *PeerDesc) GetLastPingMicros() int64 {
	if m != nil {
		return m.LastPingMicros
	}
	return 0
}

type WalletStats struct {
	MaintenanceInProgress       bool     `protobuf:"varint,1,opt,name=maintenance_in_progress,json=maintenanceInProgress,proto3" json:"maintenance_in_progress,omitempty"`
	MaintenanceName             string   `protobuf:"bytes,2,opt,name=maintenance_name,json=maintenanceName,proto3" json:"maintenance_name,omitempty"`
	MaintenanceCycles           int32    `protobuf:"varint,3,opt,name=maintenance_cycles,json=maintenanceCycles,proto3" json:"maintenance_cycles,omitempty"`
	MaintenanceLastBlockVisited int32    `protobuf:"varint,4,opt,name=maintenance_last_block_visited,json=maintenanceLastBlockVisited,proto3" json:"maintenance_last_block_visited,omitempty"`
	TimeOfLastMaintenance       string   `protobuf:"bytes,5,opt,name=time_of_last_maintenance,json=timeOfLastMaintenance,proto3" json:"time_of_last_maintenance,omitempty"`
	Syncing                     bool     `protobuf:"varint,6,opt,name=syncing,proto3" json:"syncing,omitempty"`
	SyncStarted                 string   `protobuf:"bytes,7,opt,name=sync_started,json=syncStarted,proto3" json:"sync_started,omitempty"`
	SyncRemainingSeconds        int64    `protobuf:"varint,8,opt,name=sync_remaining_seconds,json=syncRemainingSeconds,proto3" json:"sync_remaining_seconds,omitempty"`
	SyncCurrentBlock            int32    `protobuf:"varint,9,opt,name=sync_current_block,json=syncCurrentBlock,proto3" json:"sync_current_block,omitempty"`
	SyncFrom                    int32    `protobuf:"varint,10,opt,name=sync_from,json=syncFrom,proto3" json:"sync_from,omitempty"`
	SyncTo                      int32    `protobuf:"varint,11,opt,name=sync_to,json=syncTo,proto3" json:"sync_to,omitempty"`
	BirthdayBlock               int32    `protobuf:"varint,12,opt,name=birthday_block,json=birthdayBlock,proto3" json:"birthday_block,omitempty"`
	XXX_NoUnkeyedLiteral        struct{} `json:"-"`
	XXX_unrecognized            []byte   `json:"-"`
	XXX_sizecache               int32    `json:"-"`
}

func (m *WalletStats) Reset()         { *m = WalletStats{} }
func (m *WalletStats) String() string { return proto.CompactTextString(m) }
func (*WalletStats) ProtoMessage()    {}
func (*WalletStats) Descriptor() ([]byte, []int) {
	return fileDescriptor_3c5f63a845a51abc, []int{2}
}

func (m *WalletStats) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_WalletStats.Unmarshal(m, b)
}
func (m *WalletStats) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_WalletStats.Marshal(b, m, deterministic)
}
func (m *WalletStats) XXX_Merge(src proto.Message) {
	xxx_messageInfo_WalletStats.Merge(m, src)
}
func (m *WalletStats) XXX_Size() int {
	return xxx_messageInfo_WalletStats.Size(m)
}
func (m *WalletStats) XXX_DiscardUnknown() {
	xxx_messageInfo_WalletStats.DiscardUnknown(m)
}

var xxx_messageInfo_WalletStats proto.InternalMessageInfo

func (m *WalletStats) GetMaintenanceInProgress() bool {
	if m != nil {
		return m.MaintenanceInProgress
	}
	return false
}

func (m *WalletStats) GetMaintenanceName() string {
	if m != nil {
		return m.MaintenanceName
	}
	return ""
}

func (m *WalletStats) GetMaintenanceCycles() int32 {
	if m != nil {
		return m.MaintenanceCycles
	}
	return 0
}

func (m *WalletStats) GetMaintenanceLastBlockVisited() int32 {
	if m != nil {
		return m.MaintenanceLastBlockVisited
	}
	return 0
}

func (m *WalletStats) GetTimeOfLastMaintenance() string {
	if m != nil {
		return m.TimeOfLastMaintenance
	}
	return ""
}

func (m *WalletStats) GetSyncing() bool {
	if m != nil {
		return m.Syncing
	}
	return false
}

func (m *WalletStats) GetSyncStarted() string {
	if m != nil {
		return m.SyncStarted
	}
	return ""
}

func (m *WalletStats) GetSyncRemainingSeconds() int64 {
	if m != nil {
		return m.SyncRemainingSeconds
	}
	return 0
}

func (m *WalletStats) GetSyncCurrentBlock() int32 {
	if m != nil {
		return m.SyncCurrentBlock
	}
	return 0
}

func (m *WalletStats) GetSyncFrom() int32 {
	if m != nil {
		return m.SyncFrom
	}
	return 0
}

func (m *WalletStats) GetSyncTo() int32 {
	if m != nil {
		return m.SyncTo
	}
	return 0
}

func (m *WalletStats) GetBirthdayBlock() int32 {
	if m != nil {
		return m.BirthdayBlock
	}
	return 0
}

func init() {
	proto.RegisterType((*Config)(nil), "lnrpc.Config")
	proto.RegisterType((*PeerDesc)(nil), "lnrpc.PeerDesc")
	proto.RegisterType((*WalletStats)(nil), "lnrpc.WalletStats")
}

func init() { proto.RegisterFile("pkt.proto", fileDescriptor_3c5f63a845a51abc) }

var fileDescriptor_3c5f63a845a51abc = []byte{
	// 984 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x95, 0xed, 0x6e, 0x1b, 0x45,
	0x14, 0x86, 0x65, 0xe7, 0xa3, 0xde, 0x89, 0x3f, 0x92, 0x69, 0xd2, 0x0c, 0x0d, 0x50, 0x13, 0x4a,
	0x6b, 0x24, 0x9a, 0x22, 0x40, 0xf0, 0x3b, 0x35, 0x45, 0xad, 0xa0, 0x25, 0x5a, 0x47, 0x41, 0xe2,
	0xcf, 0x6a, 0x3c, 0x7b, 0x6c, 0x8f, 0xbc, 0x3b, 0x63, 0xcd, 0x4c, 0x9c, 0xe6, 0x5a, 0xb8, 0x15,
	0x2e, 0x83, 0x0b, 0x42, 0xe7, 0xcc, 0xae, 0xbd, 0xa0, 0xfe, 0xb0, 0xb4, 0xfb, 0xbc, 0xef, 0x7c,
	0xec, 0x9c, 0x77, 0x8e, 0x59, 0xb2, 0x5a, 0x86, 0x8b, 0x95, 0xb3, 0xc1, 0xf2, 0xbd, 0xc2, 0xb8,
	0x95, 0x3a, 0xff, 0xa7, 0xcd, 0xf6, 0xc7, 0xd6, 0xcc, 0xf4, 0x9c, 0x3f, 0x63, 0x83, 0x5b, 0x0f,
	0x2e, 0x93, 0x73, 0x30, 0x21, 0x33, 0xb2, 0x04, 0xb1, 0x33, 0x6c, 0x8d, 0x92, 0xb4, 0x87, 0xf8,
	0x12, 0xe9, 0x7b, 0x59, 0x02, 0xff, 0x86, 0xf1, 0x86, 0x6f, 0x0d, 0xce, 0x6b, 0x6b, 0xc4, 0x2e,
	0x59, 0x0f, 0x37, 0xd6, 0x9b, 0xc8, 0xf9, 0x05, 0x7b, 0xd8, 0x70, 0x2b, 0x5b, 0x96, 0x60, 0x82,
	0x17, 0x7b, 0x64, 0x3f, 0xda, 0xd8, 0xc7, 0x95, 0xc0, 0x1f, 0xb3, 0x8e, 0x07, 0xb7, 0xd6, 0x0a,
	0xbc, 0x78, 0x40, 0xa6, 0xcd, 0x3b, 0xff, 0x9a, 0x1d, 0xd2, 0xe6, 0x95, 0x2d, 0x36, 0xeb, 0x76,
	0x86, 0xad, 0x51, 0x2f, 0x1d, 0xd4, 0xbc, 0x5e, 0x76, 0xc4, 0x0e, 0x73, 0xed, 0xe5, 0xb4, 0x80,
	0xcc, 0x41, 0x21, 0xef, 0xb3, 0xf0, 0x41, 0x24, 0xc3, 0xd6, 0xa8, 0x93, 0xf6, 0x2b, 0x9e, 0x22,
	0xbe, 0xfe, 0x80, 0x93, 0x06, 0xa7, 0xd5, 0xb2, 0x80, 0x4c, 0x9b, 0x00, 0x6e, 0x2d, 0x0b, 0x71,
	0x30, 0x6c, 0x8d, 0x76, 0xd2, 0x41, 0xc5, 0xdf, 0x56, 0x18, 0x27, 0x95, 0x45, 0x61, 0xef, 0x32,
	0x0f, 0xc5, 0x2c, 0x53, 0xd6, 0x18, 0x2f, 0xba, 0x71, 0x52, 0xe2, 0x13, 0x28, 0x66, 0x63, 0xa4,
	0xe7, 0x7f, 0x3f, 0x60, 0x9d, 0x2b, 0x00, 0xf7, 0x33, 0x78, 0xc5, 0xbf, 0x62, 0xfd, 0xe9, 0x7d,
	0x00, 0x9f, 0x39, 0x50, 0xa0, 0xd7, 0x90, 0x8b, 0xd6, 0xb0, 0x35, 0xda, 0x4d, 0x7b, 0x44, 0xd3,
	0x0a, 0xf2, 0xcf, 0x18, 0x8b, 0x36, 0x0f, 0x26, 0x88, 0x36, 0x59, 0x12, 0x22, 0x13, 0x30, 0x81,
	0x9f, 0xb1, 0xa4, 0x90, 0x3e, 0xe0, 0x24, 0xeb, 0xaa, 0x30, 0x1d, 0x04, 0x29, 0xa8, 0xf5, 0x46,
	0xf4, 0x60, 0xf2, 0xaa, 0x14, 0x24, 0x4e, 0xc0, 0xe4, 0xfc, 0x53, 0x96, 0xe0, 0x5e, 0x41, 0x05,
	0xc8, 0xe9, 0xe0, 0x3b, 0xe9, 0x16, 0x70, 0xce, 0x76, 0x65, 0x9e, 0x3b, 0xb1, 0x4f, 0xa3, 0xe8,
	0x99, 0x3f, 0x61, 0x3b, 0x6a, 0x36, 0xa7, 0xf3, 0x3f, 0xf8, 0xae, 0x77, 0x41, 0x51, 0xb9, 0x88,
	0x31, 0x49, 0x51, 0xe1, 0x82, 0x3d, 0xd0, 0x66, 0x6a, 0x6f, 0x4d, 0x4e, 0x05, 0xe8, 0xa4, 0xf5,
	0x2b, 0xef, 0xb3, 0xb6, 0x91, 0x74, 0xd4, 0x49, 0xda, 0x36, 0x12, 0xdf, 0x75, 0x2e, 0xd8, 0xb0,
	0x35, 0xda, 0x4b, 0xdb, 0x9a, 0xbe, 0x72, 0x9b, 0x07, 0x3a, 0xe8, 0x24, 0x4d, 0x36, 0x31, 0xf8,
	0x4f, 0xf9, 0xbb, 0xff, 0x2b, 0xff, 0x97, 0xac, 0x57, 0x55, 0x3d, 0x5b, 0x1a, 0x7b, 0x67, 0x44,
	0x8f, 0x96, 0xee, 0x56, 0xf0, 0x57, 0x64, 0xfc, 0x5b, 0x76, 0x2c, 0xf3, 0x35, 0xb8, 0xa0, 0x3d,
	0xe4, 0x19, 0xc5, 0x02, 0xb3, 0x22, 0xfa, 0x94, 0x13, 0xbe, 0xd5, 0xae, 0x50, 0xba, 0x01, 0xf7,
	0xd1, 0x54, 0x0d, 0x3e, 0x9e, 0xaa, 0x1f, 0xd8, 0x23, 0x3c, 0xe1, 0x6c, 0x01, 0x32, 0x07, 0xe7,
	0xb3, 0x95, 0x83, 0x19, 0x38, 0x07, 0xb9, 0x38, 0xa4, 0xad, 0x1c, 0xa3, 0xfa, 0x26, 0x8a, 0x57,
	0xb5, 0x86, 0xb1, 0x59, 0xe3, 0x17, 0xab, 0xe5, 0x36, 0x01, 0x47, 0x31, 0x36, 0x6b, 0x70, 0x97,
	0x6a, 0xb9, 0x89, 0xc0, 0x73, 0x36, 0xb8, 0xd3, 0xc1, 0x80, 0xf7, 0x19, 0x18, 0x0c, 0x69, 0x2e,
	0x78, 0x34, 0x56, 0xf8, 0x75, 0xa4, 0x78, 0x14, 0x77, 0xda, 0x41, 0x06, 0x46, 0xd9, 0x5c, 0x9b,
	0xb9, 0x78, 0x48, 0x67, 0xd5, 0x45, 0xf8, 0xba, 0x62, 0xfc, 0x09, 0x3b, 0x08, 0xba, 0x84, 0xcc,
	0xce, 0x66, 0x1e, 0x82, 0x38, 0xa6, 0x50, 0x33, 0x44, 0xbf, 0x13, 0xc1, 0x60, 0x92, 0x61, 0x9b,
	0x8e, 0x93, 0x78, 0xe1, 0x91, 0x8e, 0x37, 0x09, 0x79, 0xce, 0x06, 0x3e, 0x48, 0x17, 0xb4, 0x99,
	0x67, 0x0b, 0xd0, 0xf3, 0x45, 0x10, 0x8f, 0xa8, 0x9e, 0xfd, 0x1a, 0xbf, 0x21, 0x8a, 0xb5, 0xa5,
	0x14, 0x4e, 0x0b, 0xab, 0x96, 0xe2, 0x94, 0x3c, 0x94, 0xcb, 0x57, 0x08, 0xb0, 0x34, 0x24, 0x4b,
	0x63, 0xec, 0xad, 0x51, 0x90, 0x57, 0x46, 0x31, 0x6c, 0x8d, 0xba, 0x29, 0x47, 0xed, 0xb2, 0x96,
	0xe2, 0x88, 0x67, 0x6c, 0x40, 0x23, 0x56, 0xb8, 0xb4, 0xb1, 0x46, 0x81, 0xf8, 0x24, 0x5e, 0x1d,
	0xc4, 0x57, 0xda, 0xcc, 0xdf, 0x23, 0xe4, 0x4f, 0x59, 0x7f, 0xeb, 0xc3, 0xcd, 0x8b, 0xc7, 0xf1,
	0x3c, 0x6a, 0xdb, 0xb5, 0x2e, 0x01, 0xeb, 0xb0, 0x75, 0x95, 0x5a, 0x39, 0xeb, 0xc5, 0x19, 0x1d,
	0x4a, 0xbf, 0xf6, 0xbd, 0x23, 0x7a, 0xfe, 0xd7, 0x2e, 0x3b, 0xf8, 0x43, 0x16, 0x05, 0x84, 0x49,
	0x90, 0xc1, 0xf3, 0x1f, 0xd9, 0x69, 0x29, 0xb1, 0x3b, 0x18, 0x69, 0x14, 0xf6, 0x09, 0x0c, 0xd6,
	0xdc, 0x81, 0xf7, 0x74, 0x95, 0x3b, 0xe9, 0x49, 0x43, 0x7e, 0x6b, 0xae, 0x2a, 0x11, 0xa3, 0xd5,
	0x1c, 0x47, 0x3d, 0xb5, 0x4d, 0x3b, 0x1b, 0x34, 0x38, 0x75, 0xd5, 0x17, 0x8c, 0x37, 0xad, 0xea,
	0x5e, 0x15, 0xe0, 0xe9, 0x9e, 0xef, 0xa5, 0x47, 0x0d, 0x65, 0x4c, 0x02, 0x1f, 0xb3, 0xcf, 0x9b,
	0xf6, 0xed, 0xb1, 0x67, 0x6b, 0xed, 0x35, 0x96, 0x72, 0x97, 0x86, 0x9e, 0x35, 0x5c, 0xbf, 0xd5,
	0x95, 0xb8, 0x89, 0x16, 0xfe, 0x13, 0x13, 0x55, 0x40, 0xe2, 0x04, 0x0d, 0x6f, 0xd5, 0xa0, 0x4f,
	0x62, 0x5a, 0x70, 0xe4, 0xbb, 0xad, 0x88, 0xd7, 0xdf, 0xdf, 0x1b, 0x85, 0xc1, 0xdb, 0x8f, 0xd7,
	0xbf, 0x7a, 0xe5, 0x5f, 0xb0, 0x2e, 0x3e, 0x66, 0x94, 0x0c, 0xc8, 0xab, 0x16, 0x7e, 0x80, 0x6c,
	0x12, 0x11, 0x5d, 0x22, 0xb4, 0x38, 0xc0, 0xf5, 0xb0, 0x16, 0x1e, 0x94, 0x35, 0xb9, 0xa7, 0x56,
	0xb2, 0x93, 0x1e, 0xa3, 0x9a, 0xd6, 0xe2, 0x24, 0x6a, 0xf8, 0xaf, 0x43, 0xa3, 0xd4, 0xad, 0x73,
	0xf8, 0x4f, 0x12, 0xa3, 0x93, 0xd0, 0x47, 0x1e, 0xa2, 0x32, 0x8e, 0x42, 0x0c, 0xce, 0x19, 0x4b,
	0xc8, 0x3d, 0x73, 0xb6, 0xac, 0x9a, 0x4f, 0x07, 0xc1, 0x2f, 0xce, 0x96, 0xfc, 0x34, 0xee, 0x3e,
	0x0b, 0x96, 0xfa, 0xcf, 0x5e, 0xba, 0x8f, 0xaf, 0xd7, 0x96, 0x1a, 0xb5, 0x76, 0x61, 0x91, 0xcb,
	0xfb, 0x6a, 0xfe, 0x2e, 0xe9, 0xbd, 0x9a, 0xd2, 0xe4, 0xaf, 0x9e, 0xfe, 0x79, 0x3e, 0xd7, 0x61,
	0x71, 0x3b, 0xbd, 0x50, 0xb6, 0x7c, 0xb9, 0x5a, 0x86, 0x17, 0x4a, 0xfa, 0x05, 0x3e, 0xe4, 0x2f,
	0x0b, 0x83, 0x3f, 0xb7, 0x52, 0xd3, 0x7d, 0x6a, 0x1e, 0xdf, 0xff, 0x1b, 0x00, 0x00, 0xff, 0xff,
	0xbf, 0x8a, 0xfd, 0x9f, 0x74, 0x07, 0x00, 0x00,
}
