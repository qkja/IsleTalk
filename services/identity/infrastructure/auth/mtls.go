package auth

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"

	"google.golang.org/grpc/credentials"
)

// ServerTLS 返回 gRPC 服务端 mTLS 凭据（T044，宪法 E 组）。
// 服务间连接 MUST 100% 启用 mTLS：要求并校验客户端证书，TLS ≥1.3。
func ServerTLS() (credentials.TransportCredentials, error) {
	certFile := os.Getenv("IDENTITY_TLS_CERT")
	keyFile := os.Getenv("IDENTITY_TLS_KEY")
	caFile := os.Getenv("IDENTITY_TLS_CA")
	if certFile == "" || keyFile == "" || caFile == "" {
		return nil, errors.New("IDENTITY_TLS_CERT / _KEY / _CA must be set for mTLS")
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	caPEM, err := os.ReadFile(caFile)
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		return nil, errors.New("failed to append CA cert")
	}

	return credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert, // MUST 校验对端证书（宪法 E 组）
		ClientCAs:    pool,
		MinVersion:   tls.VersionTLS13,
	}), nil
}

// ClientTLS 返回 gRPC 客户端 mTLS 凭据（供 relationship/message 等服务调用本服务）。
func ClientTLS(serverName string) (credentials.TransportCredentials, error) {
	certFile := os.Getenv("IDENTITY_TLS_CERT")
	keyFile := os.Getenv("IDENTITY_TLS_KEY")
	caFile := os.Getenv("IDENTITY_TLS_CA")
	if certFile == "" || keyFile == "" || caFile == "" {
		return nil, errors.New("IDENTITY_TLS_CERT / _KEY / _CA must be set for mTLS")
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	caPEM, err := os.ReadFile(caFile)
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		return nil, errors.New("failed to append CA cert")
	}

	return credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      pool,
		ServerName:   serverName, // 校验对端 SAN 与服务身份（宪法 E 组）
		MinVersion:   tls.VersionTLS13,
	}), nil
}
