# Dependency Tracker

## Overview

The Dependency Tracker is a tool designed to process dependency data from POM (Project Object Model) files. It facilitates the retrieval of transitive dependencies for each POM dependency recursively up to a level specified in the \`.env\` file. This tool offers insights into the structure of project dependencies, aiding in dependency management and understanding.

## Features

- **Dependency Processing**: The Dependency Tracker processes dependency data from POM files, extracting essential information such as group ID, artifact ID, and version.
  
- **Transitive Dependency Resolution**: It fetches transitive dependencies for each POM dependency, allowing for a comprehensive understanding of dependency chains within a project.

- **Customizable Depth**: The tool allows users to specify the recursion depth for fetching transitive dependencies, enabling fine-grained control over the analysis.

- **Input by Repository URL**: The Dependency Tracker takes a repository URL as user input, enabling the analysis of dependencies for specific projects.

## Getting Started

### Prerequisites

- Go (Golang)

### Installation

1. Clone the repository:

   \`\`\`bash
   git clone https://github.com/chhavituteja/dependency-tracker.git
   cd dependency-tracker
   \`\`\`

2. Configure \`.env\` file with desired settings:

   \`\`\`plaintext
   RECURSION_DEPTH=3
   \`\`\`

### Usage

1. Run the tool:

   \`\`\`bash
   go run main.go
   \`\`\`

2. Enter the repository URL when prompted.

3. View the processed dependency data and transitive dependencies.


## Configuration

The \`.env\` file allows for customization of the tool's behavior. Available configuration options include:

- \`RECURSION_DEPTH\`: Specifies the depth for fetching transitive dependencies recursively. Default value is 3.
