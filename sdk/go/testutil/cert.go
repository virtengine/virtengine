package testutil

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	certsmocks "github.com/virtengine/virtengine/sdk/go/mocks/node/client/cert"
	types "github.com/virtengine/virtengine/sdk/go/node/cert/v1"
	certutils "github.com/virtengine/virtengine/sdk/go/node/cert/v1/utils"
)

type TestCertificate struct {
	Cert   []tls.Certificate
	PubKey crypto.PublicKey
	Serial big.Int
	PEM    struct {
		Cert []byte
		Priv []byte
		Pub  []byte
	}
}

type certificateOption struct {
	domains []string
	nbf     time.Time
	naf     time.Time
	qclient *certsmocks.QueryClient
	ccache  CertCache
}

type CertCache interface {
	AddAccountCertificate(ctx context.Context, addr sdk.Address, cert *x509.Certificate, pubkey crypto.PublicKey) error
}

type CertificateOption func(*certificateOption)

func CertificateOptionDomains(domains []string) CertificateOption {
	return func(opt *certificateOption) {
		opt.domains = domains
	}
}

func CertificateOptionNotBefore(tm time.Time) CertificateOption {
	return func(opt *certificateOption) {
		opt.nbf = tm
	}
}

func CertificateOptionNotAfter(tm time.Time) CertificateOption {
	return func(opt *certificateOption) {
		opt.naf = tm
	}
}

func CertificateOptionMocks(val *certsmocks.QueryClient) CertificateOption {
	return func(opt *certificateOption) {
		opt.qclient = val
	}
}

func CertificateOptionCache(val CertCache) CertificateOption {
	return func(opt *certificateOption) {
		opt.ccache = val
	}
}

func Certificate(t testing.TB, issuer sdk.Address, opts ...CertificateOption) TestCertificate {
	t.Helper()

	opt := &certificateOption{}

	for _, o := range opts {
		o(opt)
	}

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	if opt.nbf.IsZero() {
		opt.nbf = time.Now()
	}

	if opt.naf.IsZero() {
		opt.naf = opt.nbf.Add(time.Hour * 24 * 365)
	}

	extKeyUsage := []x509.ExtKeyUsage{
		x509.ExtKeyUsageClientAuth,
	}

	if len(opt.domains) != 0 {
		extKeyUsage = append(extKeyUsage, x509.ExtKeyUsageServerAuth)
	}

	template := x509.Certificate{
		Subject: pkix.Name{
			CommonName: issuer.String(),
			ExtraNames: []pkix.AttributeTypeAndValue{
				{
					Type:  certutils.AuthVersionOID,
					Value: "v0.0.1",
				},
			},
		},
		Issuer: pkix.Name{
			CommonName: issuer.String(),
		},
		NotBefore:             opt.nbf,
		NotAfter:              opt.naf,
		KeyUsage:              x509.KeyUsageDataEncipherment | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           extKeyUsage,
		BasicConstraintsValid: true,
	}

	var ips []net.IP

	for i := len(opt.domains) - 1; i >= 0; i-- {
		if ip := net.ParseIP(opt.domains[i]); ip != nil {
			ips = append(ips, ip)
			opt.domains = append(opt.domains[:i], opt.domains[i+1:]...)
		}
	}

	if len(opt.domains) != 0 || len(ips) != 0 {
		template.PermittedDNSDomainsCritical = true
		template.PermittedDNSDomains = opt.domains
		template.DNSNames = opt.domains
		template.IPAddresses = ips
	}

	var certDer []byte
	if certDer, err = x509.CreateCertificate(rand.Reader, &template, &template, priv.Public(), priv); err != nil {
		t.Fatal(err)
	}

	var keyDer []byte
	if keyDer, err = x509.MarshalPKCS8PrivateKey(priv); err != nil {
		t.Fatal(err)
	}

	var pubKeyDer []byte
	if pubKeyDer, err = x509.MarshalPKIXPublicKey(priv.Public()); err != nil {
		t.Fatal(err)
	}

	x509Cert, _ := x509.ParseCertificate(certDer)

	res := TestCertificate{
		Serial: *x509Cert.SerialNumber,
		PubKey: priv.PublicKey,
		PEM: struct {
			Cert []byte
			Priv []byte
			Pub  []byte
		}{
			Cert: pem.EncodeToMemory(&pem.Block{
				Type:  types.PemBlkTypeCertificate,
				Bytes: certDer,
			}),
			Priv: pem.EncodeToMemory(&pem.Block{
				Type:  types.PemBlkTypeECPrivateKey,
				Bytes: keyDer,
			}),
			Pub: pem.EncodeToMemory(&pem.Block{
				Type:  types.PemBlkTypeECPublicKey,
				Bytes: pubKeyDer,
			}),
		},
	}

	cert, err := tls.X509KeyPair(res.PEM.Cert, res.PEM.Priv)
	if err != nil {
		t.Fatal(err)
	}

	res.Cert = append(res.Cert, cert)

	if opt.qclient != nil {
		opt.qclient.On("Certificates",
			mock.Anything,
			&types.QueryCertificatesRequest{
				Filter: types.CertificateFilter{
					Owner:  issuer.String(),
					Serial: res.Serial.String(),
					State:  types.CertificateValid.String(),
				},
			}).
			Return(&types.QueryCertificatesResponse{
				Certificates: types.CertificatesResponse{
					types.CertificateResponse{
						Certificate: types.Certificate{
							State:  types.CertificateValid,
							Cert:   res.PEM.Cert,
							Pubkey: res.PEM.Pub,
						},
						Serial: res.Serial.String(),
					},
				},
			}, nil)
	}

	if opt.ccache != nil {
		err = opt.ccache.AddAccountCertificate(context.TODO(), issuer, x509Cert, priv.Public())
		if err != nil {
			t.Fatal(err)
		}
	}

	return res
}

func CertificateRequireEqualResponse(t *testing.T, cert TestCertificate, resp types.CertificateResponse, state types.State) {
	t.Helper()

	require.Equal(t, state, resp.Certificate.State)
	require.Equal(t, cert.PEM.Cert, resp.Certificate.Cert)
	require.Equal(t, cert.PEM.Pub, resp.Certificate.Pubkey)
}
