package main

type DependencyType struct {
	Dependencies []struct {
		Dependency []struct {
			GroupId    string `xml:"groupId"`
			ArtifactId string `xml:"artifactId"`
			Version    string `xml:"version"`
		} `xml:"dependency"`
	} `xml:"dependencies"`
}

type DependencyData struct {
	URL        string `json:"dependencyPurl"`
	GroupId    string `json:"dependencyNamespace"`
	ArtifactId string `json:"dependencyName"`
	Version    string `json:"dependencyVersion"`
}

type APIResponse struct {
	Components []DependencyData `json:"components"`
	Page       int              `json:"page"`
	PageCount  int              `json:"pageCount"`
}

type Node struct {
	Id     string
	Parent *Node
}
