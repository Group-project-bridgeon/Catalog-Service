package postgres

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"net"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/domain"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/router"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/security"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func TestStockE2ECatalogHost(t *testing.T) {
	if os.Getenv("STOCK_E2E_HOST") != "1" {
		t.Skip("requires isolated parent harness")
	}
	gin.SetMode(gin.ReleaseMode)
	decoder := json.NewDecoder(os.Stdin)
	var pub string
	if decoder.Decode(&pub) != nil {
		t.Fatal("fixture public key input")
	}
	path := filepath.Join(t.TempDir(), "public.pem")
	if os.WriteFile(path, []byte(pub), 0600) != nil {
		t.Fatal("fixture public key")
	}
	machine, err := security.LoadMachine(path, "erp-auth")
	if err != nil {
		t.Fatal("fixture verifier")
	}
	f := catalogStockDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	f.ctx = ctx
	f.eligible(t)
	pack, err := f.s.CreatePack(ctx, f.variant, "BOX6", "Six pieces", 6, true)
	catalogStockOK(t, err)
	piece := f.resolve(t, domain.StockReferenceRequest{VariantID: f.variant})
	packed := f.resolve(t, domain.StockReferenceRequest{VariantID: f.variant, PackID: pack.ID, PackRevision: "1"})
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal("fixture key")
	}
	cert := &x509.Certificate{SerialNumber: big.NewInt(time.Now().UnixNano()), Subject: pkix.Name{CommonName: "isolated Catalog"}, NotBefore: time.Now().Add(-time.Minute), NotAfter: time.Now().Add(time.Hour), IPAddresses: []net.IP{net.ParseIP("127.0.0.1")}, KeyUsage: x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, IsCA: true, BasicConstraintsValid: true}
	der, err := x509.CreateCertificate(rand.Reader, cert, cert, &key.PublicKey, key)
	if err != nil {
		t.Fatal("fixture certificate")
	}
	r := router.New(f.pool, zerolog.Nop(), 1<<20, time.Second, nil, router.StockReferenceTransport{Verifier: machine, Resolver: f.s})
	server := httptest.NewUnstartedServer(r)
	server.TLS = &tls.Config{MinVersion: tls.VersionTLS12, Certificates: []tls.Certificate{{Certificate: [][]byte{der}, PrivateKey: key}}}
	server.StartTLS()
	defer server.Close()
	bootstrap := struct{ URL, CA, Variant, Pack, PieceRevision, PackReferenceRevision string }{server.URL + router.StockReferencePath, string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})), f.variant.String(), pack.ID.String(), piece.Reference.Revision, packed.Reference.Revision}
	if json.NewEncoder(os.Stdout).Encode(bootstrap) != nil {
		t.Fatal("fixture pipe")
	}
	for {
		var command string
		if decoder.Decode(&command) != nil || command == "stop" {
			break
		}
		switch command {
		case "rename":
			f.exec(t, "UPDATE catalog_products SET name='Changed synthetic product' WHERE id=$1", f.product)
		case "inactive":
			f.exec(t, "UPDATE catalog_products SET is_active=false WHERE id=$1", f.product)
		case "obsolete":
			_, err = f.s.RevisePack(ctx, pack.ID, 1, 12, "Twelve pieces")
			catalogStockOK(t, err)
		default:
			t.Fatal("unknown fixture command")
		}
		if json.NewEncoder(os.Stdout).Encode(true) != nil {
			t.Fatal("fixture acknowledgement")
		}
	}
}
