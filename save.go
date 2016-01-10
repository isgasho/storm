package storm

import (
	"encoding/json"
	"errors"

	"github.com/boltdb/bolt"
	"github.com/fatih/structs"
)

// Save a structure
func (s *Storm) Save(data interface{}) error {
	if !structs.IsStruct(data) {
		return errors.New("provided data must be a struct or a pointer to struct")
	}

	t, err := extractTags(data)
	if err != nil {
		return err
	}

	if t.ID == nil {
		if t.IDField == nil {
			return errors.New("missing struct tag id")
		}
		t.ID = t.IDField
	}

	id, err := toBytes(t.ID)
	if err != nil {
		return err
	}

	err = s.Bolt.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(t.Name))
		if err != nil {
			return err
		}

		if t.Unique != nil {
			for _, field := range t.Unique {
				key, err := toBytes(field.Value())
				if err != nil {
					return err
				}

				err = s.addToUniqueIndex([]byte(field.Name()), id, key, bucket)
				if err != nil {
					return err
				}
			}
		}

		raw, err := json.Marshal(data)
		if err != nil {
			return err
		}

		return bucket.Put(id, raw)
	})
	return err
}
