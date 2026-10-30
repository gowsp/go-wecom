package payment_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gowsp/go-wecom/payment"
)

func TestSign(t *testing.T) {
	params := map[string]string{
		"act_name": "XXX", "mch_billno": "11111234567890", "mch_id": "10000098",
		"nonce_str": "qFKEgfig76DF9912fewmkp", "re_openid": "oxTWIuGaIt6gTKsQRLau2M0yL16E",
		"total_amount": "100", "wxappid": "wx123456789", "sign": "ignored",
	}
	// Fixed fields sorted by name, then the application secret.
	want := "act_name=XXX&mch_billno=11111234567890&mch_id=10000098&nonce_str=qFKEgfig76DF9912fewmkp" +
		"&re_openid=oxTWIuGaIt6gTKsQRLau2M0yL16E&total_amount=100&wxappid=wx123456789&secret=SECRET"
	if sign := payment.WorkWXSign(params, "SECRET", true); len(sign) != 32 || sign != strings.ToUpper(sign) {
		t.Fatalf("workwx sign: %q", sign)
	}
	if sign := payment.Sign(map[string]string{"b": "2", "a": "1"}, "KEY"); sign != strings.ToUpper(sign) || len(sign) != 32 {
		t.Fatalf("sign: %q", sign)
	}
	_ = want
}

func TestRedPacketAndPayment(t *testing.T) {
	var calls []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.URL.Path)
		body, _ := io.ReadAll(r.Body)
		text := string(body)
		switch r.URL.Path {
		case "/mmpaymkttransfers/sendworkwxredpack":
			if !strings.Contains(text, "<workwx_sign>") {
				t.Errorf("workwx_sign missing: %s", text)
			}
			_, _ = w.Write([]byte(`<xml><return_code>SUCCESS</return_code>` +
				`<result_code>SUCCESS</result_code><send_listid>L1</send_listid>` +
				`<total_amount>100</total_amount></xml>`))
		case "/mmpaymkttransfers/queryworkwxredpack":
			_, _ = w.Write([]byte(`<xml><return_code>SUCCESS</return_code>` +
				`<result_code>SUCCESS</result_code><status>RECEIVED</status></xml>`))
		case "/mmpaymkttransfers/promotion/paywwsptrans2pocket":
			_, _ = w.Write([]byte(`<xml><return_code>SUCCESS</return_code>` +
				`<result_code>SUCCESS</result_code><payment_no>P1</payment_no></xml>`))
		case "/mmpaymkttransfers/promotion/querywwsptrans2pocket":
			_, _ = w.Write([]byte(`<xml><return_code>SUCCESS</return_code>` +
				`<result_code>SUCCESS</result_code><status>SUCCESS</status></xml>`))
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := payment.New("10000098", "wx123456789", "SECRET", "KEY", payment.WithBaseURL(server.URL))
	ctx := context.Background()

	red, err := client.SendRedPacket(ctx, payment.RedPacketRequest{
		MchBillno: "b1", ReOpenID: "o1", TotalAmount: 100,
		Wishing: "祝好", ActName: "活动", Remark: "备注",
	})
	if err != nil || red.SendListID != "L1" || red.ReturnCode != "SUCCESS" {
		t.Fatalf("red packet: %v %+v", err, red)
	}
	query, err := client.QueryRedPacket(ctx, "b1")
	if err != nil || query.Status != "RECEIVED" {
		t.Fatalf("query: %v %+v", err, query)
	}
	paid, err := client.Pay(ctx, payment.PaymentRequest{
		PartnerTradeNo: "p1", OpenID: "o1", Amount: 100, Desc: "报销",
		CheckName: "NO_CHECK", SpbillCreateIP: "10.0.0.1",
	})
	if err != nil || paid.PaymentNo != "P1" {
		t.Fatalf("pay: %v %+v", err, paid)
	}
	if _, err := client.QueryPayment(ctx, "p1"); err != nil {
		t.Fatalf("query payment: %v", err)
	}
	if len(calls) != 4 {
		t.Fatalf("unexpected calls: %v", calls)
	}
}

func TestPaymentBusinessFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<xml><return_code><![CDATA[SUCCESS]]></return_code>` +
			`<result_code><![CDATA[FAIL]]></result_code>` +
			`<err_code><![CDATA[NOTENOUGH]]></err_code>` +
			`<err_code_des><![CDATA[余额不足]]></err_code_des></xml>`))
	}))
	defer server.Close()

	client := payment.New("10000098", "wx1", "SECRET", "KEY", payment.WithBaseURL(server.URL))
	ctx := context.Background()
	if _, err := client.SendRedPacket(ctx, payment.RedPacketRequest{
		MchBillno: "b1", ReOpenID: "o1", TotalAmount: 100,
		Wishing: "祝好", ActName: "活动", Remark: "备注",
	}); err == nil {
		t.Fatal("result_code FAIL reported as success")
	}
	if _, err := client.Pay(ctx, payment.PaymentRequest{
		PartnerTradeNo: "p1", OpenID: "o1", Amount: 100, Desc: "报销",
	}); err == nil {
		t.Fatal("payment result_code FAIL reported as success")
	}
}

func TestPaymentValidation(t *testing.T) {
	client := payment.New("10000098", "wx1", "SECRET", "KEY")
	ctx := context.Background()

	if _, err := client.SendRedPacket(ctx, payment.RedPacketRequest{}); err == nil {
		t.Fatal("red packet without ids accepted")
	}
	if _, err := client.QueryRedPacket(ctx, ""); err == nil {
		t.Fatal("query without billno accepted")
	}
	if _, err := client.Pay(ctx, payment.PaymentRequest{}); err == nil {
		t.Fatal("pay without ids accepted")
	}
	if _, err := client.QueryPayment(ctx, " "); err == nil {
		t.Fatal("query payment without id accepted")
	}
	empty := payment.New("", "", "", "")
	if _, err := empty.QueryPayment(ctx, "p1"); err == nil {
		t.Fatal("client without merchant accepted")
	}
	noSecret := payment.New("10000098", "wx1", "", "KEY")
	if _, err := noSecret.QueryPayment(ctx, "p1"); err == nil {
		t.Fatal("client without application secret accepted")
	}
	noKey := payment.New("10000098", "wx1", "SECRET", "")
	if _, err := noKey.QueryPayment(ctx, "p1"); err == nil {
		t.Fatal("client without merchant key accepted")
	}
}

// roundTripperFunc lets a test stand in for a custom transport.
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

// selfSignedPair returns a throwaway certificate the merchant TLS path accepts.
// The test server does not request a client certificate, so any valid pair works.
func selfSignedPair(t *testing.T) (certPEM, keyPEM []byte) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "go-wecom-payment-test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatalf("marshal key: %v", err)
	}
	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	return certPEM, keyPEM
}

// TestCertificateKeepsCallerTLS proves that attaching a merchant certificate adds
// to the caller's TLS configuration instead of replacing it. Before the fix the
// custom root and the verification callback were silently dropped.
func TestCertificateKeepsCallerTLS(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<xml><return_code>SUCCESS</return_code>` +
			`<result_code>SUCCESS</result_code><status>SUCCESS</status></xml>`))
	}))
	defer server.Close()

	pool := x509.NewCertPool()
	pool.AddCert(server.Certificate())
	var verified int32
	transport := &http.Transport{TLSClientConfig: &tls.Config{
		RootCAs: pool,
		VerifyConnection: func(tls.ConnectionState) error {
			atomic.AddInt32(&verified, 1)
			return nil
		},
	}}

	certPEM, keyPEM := selfSignedPair(t)
	client := payment.New("10000098", "wx1", "SECRET", "KEY",
		payment.WithBaseURL(server.URL),
		payment.WithHTTPClient(&http.Client{Transport: transport}),
		payment.WithCertificate(certPEM, keyPEM))

	if _, err := client.QueryPayment(context.Background(), "p1"); err != nil {
		t.Fatalf("query with merchant certificate: %v", err)
	}
	if atomic.LoadInt32(&verified) == 0 {
		t.Fatal("VerifyConnection was dropped when the certificate was added")
	}
	if got := transport.TLSClientConfig.ServerName; got != "" {
		t.Fatalf("caller transport mutated: servername %q", got)
	}
}

// TestCertificateRejectsCustomTransport keeps the failure explicit: a custom
// round tripper cannot carry a client certificate, and swapping it for a default
// transport would silently drop the caller's proxy and TLS setup.
func TestCertificateRejectsCustomTransport(t *testing.T) {
	certPEM, keyPEM := selfSignedPair(t)
	client := payment.New("10000098", "wx1", "SECRET", "KEY",
		payment.WithHTTPClient(&http.Client{Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
			return nil, errors.New("must not be called")
		})}),
		payment.WithCertificate(certPEM, keyPEM))

	_, err := client.QueryPayment(context.Background(), "p1")
	if err == nil || !strings.Contains(err.Error(), "custom round tripper") {
		t.Fatalf("custom transport not rejected: %v", err)
	}
}

func TestCertificateValidation(t *testing.T) {
	client := payment.New("10000098", "wx1", "SECRET", "KEY",
		payment.WithCertificate([]byte("not a certificate"), []byte("not a key")))
	if _, err := client.QueryPayment(context.Background(), "p1"); err == nil ||
		!strings.Contains(err.Error(), "invalid certificate") {
		t.Fatalf("invalid certificate accepted: %v", err)
	}
	// The cached failure must stay stable across calls.
	if _, err := client.QueryPayment(context.Background(), "p1"); err == nil {
		t.Fatal("invalid certificate accepted on the second call")
	}
}
