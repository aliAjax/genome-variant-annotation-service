package storage

import "github.com/example/genome-variant-annotation/internal/platform"

type ManifestState string

const (
	ManifestBuilding ManifestState = "building"
	ManifestSealed   ManifestState = "sealed"
)

type ManifestPart struct {
	Name   string `json:"name"`
	Digest string `json:"digest"`
}

type Manifest struct {
	ID    string                   `json:"id"`
	State ManifestState            `json:"state"`
	Parts map[string]*ManifestPart `json:"parts"`
	Order []string                 `json:"order"`
}

func NewManifest(id string) *Manifest {
	return &Manifest{ID: id, State: ManifestBuilding}
}

func (m *Manifest) AddPart(part *ManifestPart) error {
	if _, exists := m.Parts[part.Name]; exists {
		return platform.ErrConflict
	}
	m.Parts[part.Name] = part
	m.Order = append(m.Order, part.Name)
	return nil
}

func cloneManifest(manifest *Manifest) *Manifest {
	if manifest == nil {
		return nil
	}
	cloned := *manifest
	cloned.Parts = make(map[string]*ManifestPart, len(manifest.Parts))
	for name, part := range manifest.Parts {
		partCopy := *part
		cloned.Parts[name] = &partCopy
	}
	cloned.Order = append([]string(nil), manifest.Order...)
	return &cloned
}
