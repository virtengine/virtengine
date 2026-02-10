// Package billing provides billing and invoice types for the escrow module.
package billing

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

// ArtifactType defines the type of invoice artifact
type ArtifactType uint8

const (
	// ArtifactTypeJSON is a JSON invoice document
	ArtifactTypeJSON ArtifactType = 0

	// ArtifactTypePDF is a PDF invoice document
	ArtifactTypePDF ArtifactType = 1

	// ArtifactTypeReceipt is a payment receipt
	ArtifactTypeReceipt ArtifactType = 2
)

// Content type constants for MIME types
const (
	contentTypeJSON        = "application/json"
	contentTypeOctetStream = "application/octet-stream"
)

// ArtifactTypeNames maps types to names
var ArtifactTypeNames = map[ArtifactType]string{
	ArtifactTypeJSON:    "json",
	ArtifactTypePDF:     "pdf",
	ArtifactTypeReceipt: "receipt",
}

// String returns string representation
func (t ArtifactType) String() string {
	if name, ok := ArtifactTypeNames[t]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", t)
}

// ContentType returns the MIME type for the artifact
func (t ArtifactType) ContentType() string {
	switch t {
	case ArtifactTypeJSON:
		return contentTypeJSON
	case ArtifactTypePDF:
		return "application/pdf"
	case ArtifactTypeReceipt:
		return contentTypeJSON
	default:
		return contentTypeOctetStream
	}
}

// InvoiceArtifact represents an invoice document artifact
type InvoiceArtifact struct {
	// CID is the content-addressable identifier
	CID string `json:"cid"`

	// InvoiceID is the invoice this artifact belongs to
	InvoiceID string `json:"invoice_id"`

	// Type is the artifact type
	Type ArtifactType `json:"type"`

	// ContentHash is the SHA-256 hash of the content
	ContentHash string `json:"content_hash"`

	// Size is the content size in bytes
	Size int64 `json:"size"`

	// StorageBackend is where the artifact is stored
	StorageBackend string `json:"storage_backend"`

	// StorageRef is the reference in the storage backend
	StorageRef string `json:"storage_ref"`

	// CreatedAt is when the artifact was created
	CreatedAt time.Time `json:"created_at"`

	// CreatedBy is who created the artifact
	CreatedBy string `json:"created_by"`
}

// Validate validates the artifact
func (a *InvoiceArtifact) Validate() error {
	if a.CID == "" {
		return fmt.Errorf("cid is required")
	}

	if a.InvoiceID == "" {
		return fmt.Errorf("invoice_id is required")
	}

	if a.ContentHash == "" {
		return fmt.Errorf("content_hash is required")
	}

	if len(a.ContentHash) != 64 {
		return fmt.Errorf("content_hash must be 64 hex characters (SHA-256)")
	}

	if a.Size <= 0 {
		return fmt.Errorf("size must be positive")
	}

	return nil
}

// ArtifactStore defines the interface for storing invoice artifacts
type ArtifactStore interface {
	// Store stores an artifact and returns its CID
	Store(ctx context.Context, invoiceID string, artifactType ArtifactType, content []byte, createdBy string) (*InvoiceArtifact, error)

	// Get retrieves an artifact by CID
	Get(ctx context.Context, cid string) ([]byte, *InvoiceArtifact, error)

	// GetByInvoice retrieves all artifacts for an invoice
	GetByInvoice(ctx context.Context, invoiceID string) ([]*InvoiceArtifact, error)

	// Delete deletes an artifact (if allowed)
	Delete(ctx context.Context, cid string) error

	// Verify verifies an artifact's content hash
	Verify(ctx context.Context, cid string, content []byte) (bool, error)
}

// MemoryArtifactStore is an in-memory implementation of ArtifactStore (for testing)
type MemoryArtifactStore struct {
	artifacts map[string]*InvoiceArtifact
	content   map[string][]byte
	byInvoice map[string][]string
}

// NewMemoryArtifactStore creates a new in-memory artifact store
func NewMemoryArtifactStore() *MemoryArtifactStore {
	return &MemoryArtifactStore{
		artifacts: make(map[string]*InvoiceArtifact),
		content:   make(map[string][]byte),
		byInvoice: make(map[string][]string),
	}
}

// Store stores an artifact
func (s *MemoryArtifactStore) Store(
	ctx context.Context,
	invoiceID string,
	artifactType ArtifactType,
	content []byte,
	createdBy string,
) (*InvoiceArtifact, error) {
	// Compute content hash
	hash := sha256.Sum256(content)
	contentHash := hex.EncodeToString(hash[:])

	// Generate CID (simplified - in production would use IPFS or similar)
	cid := fmt.Sprintf("baf%s", contentHash[:32])

	artifact := &InvoiceArtifact{
		CID:            cid,
		InvoiceID:      invoiceID,
		Type:           artifactType,
		ContentHash:    contentHash,
		Size:           int64(len(content)),
		StorageBackend: "memory",
		StorageRef:     cid,
		CreatedAt:      time.Now().UTC(),
		CreatedBy:      createdBy,
	}

	s.artifacts[cid] = artifact
	s.content[cid] = content
	s.byInvoice[invoiceID] = append(s.byInvoice[invoiceID], cid)

	return artifact, nil
}

// Get retrieves an artifact
func (s *MemoryArtifactStore) Get(ctx context.Context, cid string) ([]byte, *InvoiceArtifact, error) {
	artifact, ok := s.artifacts[cid]
	if !ok {
		return nil, nil, fmt.Errorf("artifact not found: %s", cid)
	}

	content, ok := s.content[cid]
	if !ok {
		return nil, nil, fmt.Errorf("artifact content not found: %s", cid)
	}

	return content, artifact, nil
}

// GetByInvoice retrieves all artifacts for an invoice
func (s *MemoryArtifactStore) GetByInvoice(ctx context.Context, invoiceID string) ([]*InvoiceArtifact, error) {
	cids := s.byInvoice[invoiceID]
	artifacts := make([]*InvoiceArtifact, 0, len(cids))

	for _, cid := range cids {
		if artifact, ok := s.artifacts[cid]; ok {
			artifacts = append(artifacts, artifact)
		}
	}

	return artifacts, nil
}

// Delete deletes an artifact
func (s *MemoryArtifactStore) Delete(ctx context.Context, cid string) error {
	artifact, ok := s.artifacts[cid]
	if !ok {
		return fmt.Errorf("artifact not found: %s", cid)
	}

	// Remove from byInvoice index
	invoiceID := artifact.InvoiceID
	cids := s.byInvoice[invoiceID]
	newCIDs := make([]string, 0, len(cids)-1)
	for _, c := range cids {
		if c != cid {
			newCIDs = append(newCIDs, c)
		}
	}
	s.byInvoice[invoiceID] = newCIDs

	delete(s.artifacts, cid)
	delete(s.content, cid)

	return nil
}

// Verify verifies an artifact's content hash
func (s *MemoryArtifactStore) Verify(ctx context.Context, cid string, content []byte) (bool, error) {
	artifact, ok := s.artifacts[cid]
	if !ok {
		return false, fmt.Errorf("artifact not found: %s", cid)
	}

	hash := sha256.Sum256(content)
	contentHash := hex.EncodeToString(hash[:])

	return artifact.ContentHash == contentHash, nil
}

// InvoiceDocumentGenerator generates invoice documents in various formats
type InvoiceDocumentGenerator struct {
	store ArtifactStore
}

// NewInvoiceDocumentGenerator creates a new document generator
func NewInvoiceDocumentGenerator(store ArtifactStore) *InvoiceDocumentGenerator {
	return &InvoiceDocumentGenerator{store: store}
}

// GenerateJSONDocument generates a JSON invoice document
func (g *InvoiceDocumentGenerator) GenerateJSONDocument(
	ctx context.Context,
	invoice *Invoice,
	createdBy string,
) (*InvoiceArtifact, error) {
	// Create the JSON document
	doc := InvoiceJSONDocument{
		Version:       "1.0",
		SchemaVersion: "virtengine/invoice/v1",
		Invoice:       invoice,
		GeneratedAt:   time.Now().UTC(),
	}

	content, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal invoice document: %w", err)
	}

	return g.store.Store(ctx, invoice.InvoiceID, ArtifactTypeJSON, content, createdBy)
}

// InvoiceJSONDocument is the JSON document structure
type InvoiceJSONDocument struct {
	// Version is the document version
	Version string `json:"version"`

	// SchemaVersion is the schema version
	SchemaVersion string `json:"schema_version"`

	// Invoice is the full invoice
	Invoice *Invoice `json:"invoice"`

	// GeneratedAt is when the document was generated
	GeneratedAt time.Time `json:"generated_at"`
}

// ComputeInvoiceHash computes a deterministic hash for an invoice
func ComputeInvoiceHash(inv *Invoice) (string, error) {
	if inv == nil {
		return "", fmt.Errorf("invoice is nil")
	}

	lineItems := canonicalLineItems(inv.LineItems)
	canonical := struct {
		InvoiceID     string              `json:"invoice_id"`
		InvoiceNumber string              `json:"invoice_number"`
		EscrowID      string              `json:"escrow_id"`
		OrderID       string              `json:"order_id"`
		LeaseID       string              `json:"lease_id"`
		Provider      string              `json:"provider"`
		Customer      string              `json:"customer"`
		Status        string              `json:"status"`
		Currency      string              `json:"currency"`
		BillingStart  int64               `json:"billing_start"`
		BillingEnd    int64               `json:"billing_end"`
		Subtotal      string              `json:"subtotal"`
		DiscountTotal string              `json:"discount_total"`
		TaxTotal      string              `json:"tax_total"`
		Total         string              `json:"total"`
		LineItems     []canonicalLineItem `json:"line_items"`
		Metadata      []metadataPair      `json:"metadata"`
	}{
		InvoiceID:     inv.InvoiceID,
		InvoiceNumber: inv.InvoiceNumber,
		EscrowID:      inv.EscrowID,
		OrderID:       inv.OrderID,
		LeaseID:       inv.LeaseID,
		Provider:      inv.Provider,
		Customer:      inv.Customer,
		Status:        inv.Status.String(),
		Currency:      inv.Currency,
		BillingStart:  inv.BillingPeriod.StartTime.Unix(),
		BillingEnd:    inv.BillingPeriod.EndTime.Unix(),
		Subtotal:      inv.Subtotal.String(),
		DiscountTotal: inv.DiscountTotal.String(),
		TaxTotal:      inv.TaxTotal.String(),
		Total:         inv.Total.String(),
		LineItems:     lineItems,
		Metadata:      sortedMetadataPairs(inv.Metadata),
	}

	data, err := json.Marshal(canonical)
	if err != nil {
		return "", fmt.Errorf("failed to marshal canonical invoice: %w", err)
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

type canonicalLineItem struct {
	LineItemID     string         `json:"line_item_id"`
	Description    string         `json:"description"`
	UsageType      string         `json:"usage_type"`
	Quantity       string         `json:"quantity"`
	Unit           string         `json:"unit"`
	UnitPrice      string         `json:"unit_price"`
	Amount         string         `json:"amount"`
	UsageRecordIDs []string       `json:"usage_record_ids"`
	Metadata       []metadataPair `json:"metadata"`
}

func canonicalLineItems(items []LineItem) []canonicalLineItem {
	if len(items) == 0 {
		return nil
	}

	canonical := make([]canonicalLineItem, 0, len(items))
	for _, item := range items {
		canonical = append(canonical, canonicalLineItem{
			LineItemID:     item.LineItemID,
			Description:    item.Description,
			UsageType:      item.UsageType.String(),
			Quantity:       item.Quantity.String(),
			Unit:           item.Unit,
			UnitPrice:      item.UnitPrice.String(),
			Amount:         item.Amount.String(),
			UsageRecordIDs: append([]string(nil), item.UsageRecordIDs...),
			Metadata:       sortedMetadataPairs(item.Metadata),
		})
	}

	sort.SliceStable(canonical, func(i, j int) bool {
		if canonical[i].LineItemID == canonical[j].LineItemID {
			return canonical[i].Description < canonical[j].Description
		}
		return canonical[i].LineItemID < canonical[j].LineItemID
	})

	return canonical
}
