package cwl

import (
	"database/sql"
	"fmt"
	"github.com/dominikbraun/graph"
	"github.com/dominikbraun/graph/draw"
	"log"
	"os"
)

// Step represents a workflow step
type step struct {
	Type     string // "manual" or "automatic"
	StepID   string
	SubSteps []string // Only for automatic steps
}

// Relationship represents a relationship between two entities
type Relationship struct {
	ID1              string
	RelationshipType string
	ID2              string
}

// ExecutionOrder represents the ordered list of steps and services
type ExecutionOrder []string

// GetWorkflowExecutionOrder retrieves the execution order of a workflow
func GetWorkflowExecutionOrder(db *sql.DB, workflowName string, path string) (graph.Graph[string, string], error) {
	// Step 1: Retrieve all steps associated with the workflow
	stepsQuery := `
        SELECT id1 AS step_id
        FROM ST_WF
        WHERE id2 = ?;
    `
	rows, err := db.Query(stepsQuery, workflowName)
	if err != nil {
		return nil, fmt.Errorf("error querying ST_WF: %v", err)
	}
	defer rows.Close()

	var steps []step
	for rows.Next() {
		var stepID string
		if err := rows.Scan(&stepID); err != nil {
			return nil, fmt.Errorf("error scanning step_id: %v", err)
		}

		// Check if the step has any associated services in SS_ST
		servicesQuery := `
            SELECT id1 AS service_id
            FROM SS_ST
            WHERE id2 = ?;
        `
		serviceRows, err := db.Query(servicesQuery, stepID)
		if err != nil {
			return nil, fmt.Errorf("error querying SS_ST for step %s: %v", stepID, err)
		}

		var services []string
		for serviceRows.Next() {
			var serviceID string
			if err := serviceRows.Scan(&serviceID); err != nil {
				serviceRows.Close()
				return nil, fmt.Errorf("error scanning service_id: %v", err)
			}
			services = append(services, serviceID)
		}
		serviceRows.Close()
		if err := serviceRows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating services for step %s: %v", stepID, err)
		}

		if len(services) == 0 {
			// Manual step
			steps = append(steps, step{
				Type:   "manual",
				StepID: stepID,
			})
		} else {
			// Automatic step with associated services
			steps = append(steps, step{
				Type:     "automatic",
				StepID:   stepID,
				SubSteps: services,
			})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating steps: %v", err)
	}

	// Collect all service IDs
	var allServices []string
	for _, step := range steps {
		if step.Type == "automatic" {
			allServices = append(allServices, step.SubSteps...)
		}
	}

	// Step 2: Retrieve all relationships
	relationships, err := getAllRelationships(db, steps, allServices)
	if err != nil {
		return nil, fmt.Errorf("error retrieving relationships: %v", err)
	}

	// Step 3: Build dependency graph
	gr, err := buildDependencyGraph(steps, relationships, path)
	if err != nil {
		return nil, fmt.Errorf("error building dependency graph: %v", err)
	}

	return gr, nil
}

// getAllRelationships retrieves all relevant relationships from the database
func getAllRelationships(db *sql.DB, steps []step, services []string) ([]Relationship, error) {
	var relationships []Relationship

	// Helper function to collect IDs for SQL IN clause
	collectIDs := func(items []string) string {
		if len(items) == 0 {
			return "''" // Empty string to prevent SQL error
		}
		query := ""
		for i, id := range items {
			if i > 0 {
				query += ","
			}
			query += fmt.Sprintf("'%s'", id)
		}
		return query
	}

	var stepIDs []string
	for _, step := range steps {
		stepIDs = append(stepIDs, step.StepID)
	}
	serviceIDs := allUniqueServices(services)

	// Retrieve ST_ST relationships
	if len(stepIDs) > 0 {
		querystSt := fmt.Sprintf(`
            SELECT id1, relationship_type, id2
            FROM ST_ST
            WHERE id1 IN (%s) OR id2 IN (%s);
        `, collectIDs(stepIDs), collectIDs(stepIDs))
		rows, err := db.Query(querystSt)
		if err != nil {
			return nil, fmt.Errorf("error querying ST_ST: %v", err)
		}
		defer rows.Close()

		for rows.Next() {
			var r Relationship
			if err := rows.Scan(&r.ID1, &r.RelationshipType, &r.ID2); err != nil {
				rows.Close()
				return nil, fmt.Errorf("error scanning ST_ST: %v", err)
			}
			relationships = append(relationships, r)
		}
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating ST_ST: %v", err)
		}
	}

	// Retrieve SS_SS relationships
	if len(serviceIDs) > 0 {
		querySS_SS := fmt.Sprintf(`
            SELECT id1, relationship_type, id2
            FROM SS_SS
            WHERE id1 IN (%s) OR id2 IN (%s);
        `, collectIDs(serviceIDs), collectIDs(serviceIDs))
		rows, err := db.Query(querySS_SS)
		if err != nil {
			return nil, fmt.Errorf("error querying SS_SS: %v", err)
		}
		defer rows.Close()

		for rows.Next() {
			var r Relationship
			if err := rows.Scan(&r.ID1, &r.RelationshipType, &r.ID2); err != nil {
				rows.Close()
				return nil, fmt.Errorf("error scanning SS_SS: %v", err)
			}
			relationships = append(relationships, r)
		}
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating SS_SS: %v", err)
		}
	}

	//Retrieve SS_ST relationships
	//if len(serviceIDs) > 0 || len(stepIDs) > 0 {
	//	querySS_ST := fmt.Sprintf(`
	//       SELECT id1, relationship_type, id2
	//       FROM SS_ST
	//       WHERE id1 IN (%s) OR id2 IN (%s);
	//   `, collectIDs(serviceIDs), collectIDs(stepIDs))
	//	rows, err := db.Query(querySS_ST)
	//	if err != nil {
	//		return nil, fmt.Errorf("error querying SS_ST: %v", err)
	//	}
	//	defer rows.Close()
	//
	//	for rows.Next() {
	//		var r Relationship
	//		if err := rows.Scan(&r.ID1, &r.RelationshipType, &r.ID2); err != nil {
	//			rows.Close()
	//			return nil, fmt.Errorf("error scanning SS_ST: %v", err)
	//		}
	//		relationships = append(relationships, r)
	//	}
	//	if err := rows.Err(); err != nil {
	//		return nil, fmt.Errorf("error iterating SS_ST: %v", err)
	//	}
	//}

	return relationships, nil
}

// allUniqueServices ensures that the list of services is unique
func allUniqueServices(services []string) []string {
	unique := make(map[string]struct{})
	for _, s := range services {
		unique[s] = struct{}{}
	}
	var result []string
	for s := range unique {
		result = append(result, s)
	}
	return result
}

func buildDependencyGraph(steps []step, relationships []Relationship, path string) (graph.Graph[string, string], error) {
	// Initialize a directed graph
	g := graph.New(graph.StringHash, graph.Directed(), graph.PreventCycles())

	// Add nodes for steps and their services
	for _, s := range steps {
		if s.Type == "automatic" {
			var tempSteps []step
			for _, subStep := range s.SubSteps {
				tempSteps = append(tempSteps, step{
					Type:   "manual",
					StepID: subStep,
				})
			}

			sg, err := buildDependencyGraph(tempSteps, relationships, path)
			if err != nil {
				return nil, fmt.Errorf("error building dependency graph for automatic step '%s': %v", s.StepID, err)
			}

			err = os.MkdirAll(path, os.ModePerm)
			if err != nil {
				log.Fatal(err)
			}
			file, err := os.Create(path + s.StepID + ".dot")
			if err != nil {
				log.Fatal(err)
			}

			_ = draw.DOT(sg, file)
			file.Close()

			// Add the subgraph to the main graph
			//err = g.AddVerticesFrom(sg)
			//if err != nil {
			//	return nil, fmt.Errorf("error adding subgraph for automatic step '%s': %v", s.StepID, err)
			//}
			//
			//err = g.AddEdgesFrom(sg)
			//if err != nil {
			//	return nil, fmt.Errorf("error adding edges from subgraph for automatic step '%s': %v", s.StepID, err)
			//}
			//continue
		}
		if s.Type == "manual" {
			if err := g.AddVertex(s.StepID, graph.VertexAttribute("colorscheme", "blues3"), graph.VertexAttribute("style", "filled"), graph.VertexAttribute("color", "2"), graph.VertexAttribute("fillcolor", "1")); err != nil {
				return nil, fmt.Errorf("error adding step vertex '%s': %v", s.StepID, err)
			}
		} else {
			if err := g.AddVertex(s.StepID, graph.VertexAttribute("colorscheme", "greens3"), graph.VertexAttribute("style", "filled"), graph.VertexAttribute("color", "2"), graph.VertexAttribute("fillcolor", "1")); err != nil {
				return nil, fmt.Errorf("error adding step vertex '%s': %v", s.StepID, err)
			}
		}
	}

	// Add edges based on relationships
	for _, rel := range relationships {
		switch rel.RelationshipType {
		case "follows", "follows from", "follows to":
			// A follows B => B must be executed before A
			// Edge: B -> A
			if err := g.AddEdge(rel.ID2, rel.ID1, graph.EdgeAttribute("relationship", rel.RelationshipType)); err != nil {
				//return nil, fmt.Errorf("error adding edge '%s' -> '%s': %v", rel.ID2, rel.ID1, err)
				//log.Printf("Warning: error adding edge '%s' -> '%s': %v", rel.ID2, rel.ID1, err)
			}
		case "parent of":
			// A is parent of B => A must be executed before B
			// Edge: A -> B
			if err := g.AddEdge(rel.ID1, rel.ID2, graph.EdgeAttribute("relationship", rel.RelationshipType)); err != nil {
				//return nil, fmt.Errorf("error adding edge '%s' -> '%s': %v", rel.ID1, rel.ID2, err)
				//log.Printf("Warning: error adding edge '%s' -> '%s': %v", rel.ID1, rel.ID2, err)
			}
		case "manages", "is manager of":
			// A manages B => A must be executed before B
			// Edge: A -> B
			if err := g.AddEdge(rel.ID1, rel.ID2, graph.EdgeAttribute("relationship", rel.RelationshipType)); err != nil {
				//return nil, fmt.Errorf("error adding edge '%s' -> '%s': %v", rel.ID1, rel.ID2, err)
				//log.Printf("Warning: error adding edge '%s' -> '%s': %v", rel.ID1, rel.ID2, err)
			}
		case "is input to":
			// A is input to B => A must be executed before B
			// Edge: A -> B
			if err := g.AddEdge(rel.ID1, rel.ID2, graph.EdgeAttribute("relationship", rel.RelationshipType)); err != nil {
				//return nil, fmt.Errorf("error adding edge '%s' -> '%s': %v", rel.ID1, rel.ID2, err)
				//log.Printf("Warning: error adding edge '%s' -> '%s': %v", rel.ID1, rel.ID2, err)
			}
		case "is previous to":
			// A is previous to B => A must be executed before B
			// Edge: A -> B
			if err := g.AddEdge(rel.ID1, rel.ID2, graph.EdgeAttribute("relationship", rel.RelationshipType)); err != nil {
				//return nil, fmt.Errorf("error adding edge '%s' -> '%s': %v", rel.ID1, rel.ID2, err)
				//log.Printf("Warning: error adding edge '%s' -> '%s': %v", rel.ID1, rel.ID2, err)
			}
		case "is part of":
			// A is part of B => A must be executed before B
			// Edge: A -> B
			if err := g.AddEdge(rel.ID1, rel.ID2, graph.EdgeAttribute("relationship", rel.RelationshipType)); err != nil {
				//return nil, fmt.Errorf("error adding edge '%s' -> '%s': %v", rel.ID1, rel.ID2, err)
				//log.Printf("Warning: error adding edge '%s' -> '%s': %v", rel.ID1, rel.ID2, err)
			}
		case "simultaneous", "parallel to":
			// A and B are simultaneous => No dependency
			// No edge is added
		default:
			// Handle unknown relationship types if necessary
			log.Printf("WARNING: Unknown relationship type '%s' between '%s' and '%s'", rel.RelationshipType, rel.ID1, rel.ID2)
		}
	}

	return g, nil
}

// topologicalSort performs topological sorting on the dependency graph
func topologicalSort(graph map[string][]string) (ExecutionOrder, error) {
	// Calculate in-degrees
	inDegree := make(map[string]int)
	for node := range graph {
		inDegree[node] = 0
	}
	for _, deps := range graph {
		for _, dep := range deps {
			inDegree[dep]++
		}
	}

	// Initialize queue with nodes having in-degree 0
	var queue []string
	for node, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, node)
		}
	}

	var order ExecutionOrder
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		order = append(order, current)

		for _, neighbor := range graph[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	return order, nil
}
