package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"os"

	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
	"github.com/joho/godotenv"
	"gopkg.in/src-d/go-git.v4"
)

func main() {
	// repoURL:=https://github.com/swapnilbamble1438/EcommerceApp.git
	var repoUrl string
	fmt.Print("Enter the RepoURL:")
	_, err := fmt.Scan(&repoUrl)
	if err != nil {
		log.Println(err)
		return
	}
	targetDir, err := FetchEnv("TARGET_DIR")
	if err != nil {
		log.Println(err)
		return
	}
	err = gitClone(repoUrl, targetDir)
	if err != nil {
		log.Println(err)
		return
	}

	//find POM Files From the project
	pomPaths, err := findPOM(targetDir)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("The list of pom files : %s", pomPaths)

	//Parse POM file to fetch dependencies
	for _, path := range pomPaths {
		firstParent := Node{path, nil}
		pomNodes, err := addPomNodes(path, &firstParent)
		if err != nil {
			log.Println(err)
			return
		}

		log.Printf("This is dependency data list %+v", pomNodes)
		var nodes []Node

		//for each POM dependency find direct dependencies
		for idx, node := range pomNodes {
			var listOfNodes []Node
			nodes = append(nodes, node)
			transitivedependency, err := AddNode(&node, &listOfNodes, 1)
			if err != nil {
				log.Println(err)
				return
			}
			nodes = append(nodes, *transitivedependency...)
			log.Println("completed for node :", idx)
		}

		//append the parent node
		nodes = append(nodes, firstParent)
		log.Print("added Nodes :: ", nodes)

		//render graph
		err = RenderGraph(&nodes)
		if err != nil {
			log.Println(err)
			return
		}
	}

}

func gitClone(repoURL string, targetDIR string) error {
	_, err := git.PlainClone(targetDIR, false, &git.CloneOptions{
		URL:      repoURL,
		Progress: os.Stdout,
	})
	return err
}

func RenderGraph(nodes *[]Node) error {
	g := graphviz.New()
	graph, err := g.Graph()
	if err != nil {
		return err
	}
	defer func() {
		if err := graph.Close(); err != nil {
			log.Fatal(err)
			return
		}
		g.Close()
	}()
	//loop through nodes to create and add nodes to the graph
	graphNodeMap := make(map[string]*cgraph.Node)
	for _, node := range *nodes {
		if _, exists := graphNodeMap[node.Id]; exists {
			log.Printf("Duplicate node ID: %s", node.Id)
			continue
		}
		graphNodeMap[node.Id], err = graph.CreateNode(node.Id)
		if err != nil {
			return err
		}
	}

	for _, node := range *nodes {
		if node.Parent != nil {
			if parent, ok := graphNodeMap[node.Parent.Id]; ok {
				if _, err := graph.CreateEdge("", parent, graphNodeMap[node.Id]); err != nil {
					log.Print("Error creating edge ")
					return err
				}
			} else {
				log.Printf("Parent node not found: %s", node.Parent.Id)
				return err
			}
		}
	}

	if err := g.RenderFilename(graph, graphviz.PNG, "./graph.png"); err != nil {
		return err
	}
	return nil
}

func findPOM(rootPath string) ([]string, error) {
	var list []string
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Name() == "pom.xml" {
			log.Println(path)
			list = append(list, path)
		}
		return nil
	})
	return list, err
}

func addPomNodes(path string, parent *Node) ([]Node, error) {
	nodes := []Node{}
	var dependencies DependencyType
	xmlData, err := os.ReadFile(path)
	if err != nil {
		return []Node{}, err
	}
	err = xml.Unmarshal(xmlData, &dependencies)
	if err != nil {
		return []Node{}, err
	}
	for _, d := range dependencies.Dependencies {
		for _, dependencydata := range d.Dependency {
			if dependencydata.GroupId == "" || dependencydata.ArtifactId == "" || dependencydata.Version == "" {
				log.Println("POM file Incorrectly formed::: ", dependencydata.GroupId+"/"+dependencydata.ArtifactId+"@"+dependencydata.Version)
			}
			Id := dependencydata.GroupId + "/" + dependencydata.ArtifactId + "@" + dependencydata.Version
			node := Node{Id, parent}
			nodes = append(nodes, node)
		}

	}
	return nodes, nil

}

// method to recursively call API to fetch All transitive dependencies for a direct POM dependency
func AddNode(node *Node, listofNodes *[]Node, counter int) (*[]Node, error) {
	threshold, err := FetchEnv("RECURSION_DEPTH")
	if err != nil {
		return nil, err
	}
	levelThreshold, err := strconv.Atoi(threshold)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	if counter > levelThreshold {
		log.Println("The counter threshold crossed returning list of nodes")
		return listofNodes, nil
	}
	client := &http.Client{}
	url, err := FetchEnv("API_URL")
	if err != nil {
		return nil, err
	}
	body, err := FetchEnv("BODY_JSON")
	if err != nil {
		return nil, err
	}
	page := 0
	counter = counter + 1
	var componentList []DependencyData
	respBody, err := callAPI(client, url, body, node.Id, page)
	if err != nil {
		return nil, err
	}
	componentList = append(componentList, respBody.Components...)
	for page < respBody.PageCount-1 {
		page = page + 1
		responseBody, err := callAPI(client, url, body, node.Id, page)
		if err != nil {
			return nil, err
		}
		componentList = append(componentList, responseBody.Components...)
	}
	if len(componentList) > 0 {
		var newNodeList []Node
		for _, component := range componentList {
			Id := component.URL[10:]
			newNodeList = append(newNodeList, Node{Id, node})

		}
		for _, newnode := range newNodeList {
			*listofNodes = append(*listofNodes, newnode)
			listofNodes, err = AddNode(&newnode, listofNodes, counter)
			if err != nil {
				return nil, err
			}
		}

	} else {
		return listofNodes, nil
	}
	return listofNodes, nil

}

// method to call API to fetch direct dependencies of a dependency
func callAPI(client *http.Client, url string, body string, id string, page int) (APIResponse, error) {
	updatedbody := fmt.Sprintf(body, id, page)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer([]byte(updatedbody)))
	if err != nil {
		return APIResponse{}, err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return APIResponse{}, nil
	}
	if resp.StatusCode != http.StatusOK {
		log.Println("http Status Code not OK :: ", resp.StatusCode)
		log.Println(ioutil.ReadAll(resp.Body))
		log.Println(resp.Header)

		//to handle too many requests
		if resp.StatusCode == 429 {
			min, err := time.ParseDuration("4m")
			if err != nil {
				return APIResponse{}, nil
			}
			time.Sleep(min)
			return callAPI(client, url, body, id, page)

		}

	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return APIResponse{}, nil
	}

	//find pagecount
	//loop 0th page

	var apiResponse APIResponse
	err = json.Unmarshal([]byte(respBody), &apiResponse)
	if err != nil {
		return APIResponse{}, nil
	}
	return apiResponse, nil
}

func FetchEnv(variable string) (string, error) {
	err := godotenv.Load("env/.env")
	if err != nil {
		return "", err
	}

	value := os.Getenv(variable)

	if value == "" {
		log.Println("The keyvalue is either empty or key doesn't exist in the env")
	}
	log.Println(value)
	return value, nil

}
