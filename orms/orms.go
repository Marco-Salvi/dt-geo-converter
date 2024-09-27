package orms

import (
	"database/sql"
	"fmt"
)

type WF struct {
	Name        string
	Description string
	Author      string
}

type ST struct {
	ID string
}

type SS struct {
	ID string
}

type DT struct {
	ID string
}

type DTSTRelationship struct {
	DTID             string
	STID             string
	RelationshipType string
}

type DTSSRelationship struct {
	DTID             string
	SSID             string
	RelationshipType string
}

func GetSTsForWF(db *sql.DB, wfName string) ([]ST, error) {
	query := `
		SELECT DISTINCT id1 AS st_id
		FROM ST_WF
		WHERE id2 = ?
	`

	rows, err := db.Query(query, wfName)
	if err != nil {
		return nil, fmt.Errorf("failed to query STs for WF: %v", err)
	}
	defer rows.Close()

	var stateTransitions []ST
	for rows.Next() {
		var st ST
		if err := rows.Scan(&st.ID); err != nil {
			return nil, fmt.Errorf("failed to scan ST row: %v", err)
		}
		stateTransitions = append(stateTransitions, st)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating ST rows: %v", err)
	}

	return stateTransitions, nil
}

func GetDTSTRelationships(db *sql.DB, stID string) ([]DTSTRelationship, error) {
	query := `
		SELECT id1, relationship_type, id2
		FROM DT_ST
		WHERE id2 = ?
	`

	rows, err := db.Query(query, stID)
	if err != nil {
		return nil, fmt.Errorf("failed to query DT-ST relationships: %v", err)
	}
	defer rows.Close()

	var relationships []DTSTRelationship
	for rows.Next() {
		var rel DTSTRelationship
		if err := rows.Scan(&rel.DTID, &rel.RelationshipType, &rel.STID); err != nil {
			return nil, fmt.Errorf("failed to scan DT-ST relationship row: %v", err)
		}
		relationships = append(relationships, rel)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating DT-ST relationship rows: %v", err)
	}

	return relationships, nil
}

func GetSSForST(db *sql.DB, stID string) ([]SS, error) {
	query := `
		SELECT id1 AS ss_id
		FROM SS_ST
		WHERE id2 = ?
	`

	rows, err := db.Query(query, stID)
	if err != nil {
		return nil, fmt.Errorf("failed to query SS for ST: %v", err)
	}
	defer rows.Close()

	var subStates []SS
	for rows.Next() {
		var ss SS
		if err := rows.Scan(&ss.ID); err != nil {
			return nil, fmt.Errorf("failed to scan SS row: %v", err)
		}
		subStates = append(subStates, ss)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating SS rows: %v", err)
	}

	return subStates, nil
}

func GetDTSSRelationshipsForST(db *sql.DB, stID string) ([]DTSSRelationship, error) {
	query := `
		SELECT DISTINCT dt_ss.id1 AS dt_id, dt_ss.id2 AS ss_id, dt_ss.relationship_type
		FROM DT_SS dt_ss
		JOIN SS_ST ss_st ON dt_ss.id2 = ss_st.id1
		WHERE ss_st.id2 = ?
		AND ss_st.relationship_type = 'is part of'
	`

	rows, err := db.Query(query, stID)
	if err != nil {
		return nil, fmt.Errorf("failed to query DT-SS relationships for ST: %v", err)
	}
	defer rows.Close()

	var relationships []DTSSRelationship
	for rows.Next() {
		var rel DTSSRelationship
		if err := rows.Scan(&rel.DTID, &rel.SSID, &rel.RelationshipType); err != nil {
			return nil, fmt.Errorf("failed to scan DT-SS relationship row: %v", err)
		}
		relationships = append(relationships, rel)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating DT-SS relationship rows: %v", err)
	}

	return relationships, nil
}

func GetDTsRelatedToSTviaSSs(db *sql.DB, stID string) ([]DT, error) {
	query := `
		SELECT DISTINCT dt_ss.id1 AS dt_id
		FROM DT_SS dt_ss
		JOIN SS_ST ss_st ON dt_ss.id2 = ss_st.id1
		WHERE ss_st.id2 = ?
	`

	rows, err := db.Query(query, stID)
	if err != nil {
		return nil, fmt.Errorf("failed to query DTs related to ST via SSs: %v", err)
	}
	defer rows.Close()

	var dataTypes []DT
	for rows.Next() {
		var dt DT
		if err := rows.Scan(&dt.ID); err != nil {
			return nil, fmt.Errorf("failed to scan DT row: %v", err)
		}
		dataTypes = append(dataTypes, dt)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating DT rows: %v", err)
	}

	return dataTypes, nil
}

func GetDTSSRelationshipsForSS(db *sql.DB, ssID string) ([]DTSSRelationship, error) {
	query := `
        SELECT id1 AS dt_id, id2 AS ss_id, relationship_type
        FROM DT_SS
        WHERE id2 = ?
    `

	rows, err := db.Query(query, ssID)
	if err != nil {
		return nil, fmt.Errorf("failed to query DT-SS relationships for SS: %v", err)
	}
	defer rows.Close()

	var relationships []DTSSRelationship
	for rows.Next() {
		var rel DTSSRelationship
		if err := rows.Scan(&rel.DTID, &rel.SSID, &rel.RelationshipType); err != nil {
			return nil, fmt.Errorf("failed to scan DT-SS relationship row: %v", err)
		}
		relationships = append(relationships, rel)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating DT-SS relationship rows: %v", err)
	}

	return relationships, nil
}
