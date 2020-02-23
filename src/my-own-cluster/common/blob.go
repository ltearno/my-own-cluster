package common

import (
	"encoding/json"
	"fmt"
)

type BlobAbstract struct {
	ContentType string `json:"content_type"`
	Length      int    `json:"length"`
}

func (o *Orchestrator) RegisterBlob(contentType string, contentBytes []byte) (string, error) {
	techID := Sha256Sum(contentBytes)

	has, err := o.db.Has([]byte(fmt.Sprintf("/blobs/byid/%s/abstract", techID)), nil)
	if err == nil && has {
		return techID, nil
	}

	abstract := &BlobAbstract{
		ContentType: contentType,
		Length:      len(contentBytes),
	}

	abstractBytes, err := json.Marshal(abstract)
	if err != nil {
		return "", err
	}

	o.db.Put([]byte(fmt.Sprintf("/blobs/byid/%s/abstract", techID)), abstractBytes, nil)
	o.db.Put([]byte(fmt.Sprintf("/blobs/byid/%s/bytes", techID)), contentBytes, nil)

	fmt.Printf("registered_blob '%s', content_type:%s, size:%d\n", techID, contentType, len(contentBytes))

	return techID, nil
}

func (o *Orchestrator) RegisterBlobWithName(name string, contentType string, contentBytes []byte) (string, error) {
	techID, err := o.RegisterBlob(contentType, contentBytes)
	if err != nil {
		return "", err
	}

	o.db.Put([]byte(fmt.Sprintf("/blobs/byname/%s", name)), []byte(techID), nil)

	fmt.Printf("registered_blob_by_name '%s', techID:%s\n", name, techID)

	return techID, nil
}

func (o *Orchestrator) GetBlobTechIDFromName(name string) (string, error) {
	techID, err := o.db.Get([]byte(fmt.Sprintf("/blobs/byname/%s", name)), nil)
	if err != nil {
		return "", err
	}

	return string(techID), nil
}

func (o *Orchestrator) GetBlobAbstractByTechID(techID string) (*BlobAbstract, error) {
	abstractBytes, err := o.db.Get([]byte(fmt.Sprintf("/blobs/byid/%s/abstract", techID)), nil)
	if err != nil {
		return nil, err
	}

	abstract := &BlobAbstract{}

	err = json.Unmarshal(abstractBytes, abstract)
	if err != nil {
		return nil, err
	}

	return abstract, nil
}

func (o *Orchestrator) GetBlobBytesByTechID(techID string) ([]byte, error) {
	contentBytes, err := o.db.Get([]byte(fmt.Sprintf("/blobs/byid/%s/bytes", techID)), nil)
	if err != nil {
		return nil, err
	}

	return contentBytes, nil
}
