# Demo Setup

## Reference Registry
Deploy the reference MCP registry:
```
oc new-project mcp-registry-ref
oc apply -f demo/registry-ref.yaml
```

Test:
```
MCP_REGISTRY=$(oc get route mcp-registry-route -oyaml | yq '.status.ingress[0].host') \
  curl "https://$MCP_REGISTRY/v0/servers" | jq
```

## [Admin] Install the operator
Create the target NS:
```
oc new-project mcp-catalog
```

Generate resources and manifests and install the operator:
```
make generate && make manifests
make install
```
Execute the operator from the console:
```
POD_NAMESPACE=mcp-catalog go run cmd/main.go
```

**TODO** Automate the following manual steps
* Install of build pipeline:
```
oc apply -f demo/pipeline.yaml -n mcp-catalog
```
* Binf `privileged` SCC to `pipeline` service account:
```
oc adm policy add-scc-to-user privileged \
  -z pipeline -n mcp-catalog
```
* Start of catalog service:
```bash
oc project mcp-catalog
cd ../mcp-catalog-apps/mcp-registry
MCP_REGISTRY_NAME=foo MCP_CATALOG_NAME=red-hat-ecosystem-mcp-catalog uv run uvicorn mcp_registry.app:app --host 0.0.0.0 --port 8000
```

## [Admin] Create global Catalog

Install a sample catalog instance:
```
oc apply -n mcp-catalog -f demo/catalog.yaml
```


## [Admin] Ingest server defs from reference MCP registry
Run the import service from the catalog service:
```
MCP_REGISTRY=$(oc get route mcp-registry-route -oyaml | yq '.status.ingress[0].host') \
  curl -X POST "localhost:8000/import?mcp_registry_source=https://$MCP_REGISTRY/v0" | jq
```

Verify the CRs:
```
oc get mcpserverdefinitions -n mcp-catalog | wc -l
```

## [Admin] Promote selected instance(s) to blueprint
Select one server definition and promote it:
```
SERVER_NAME="$(oc get mcpserverdefinitions -n mcp-catalog | grep tavily-mcp | head -1 | cut -d ' ' -f 1)" \
  && echo "$SERVER_NAME" \
  && curl -X POST "localhost:8000/promote?server_definition_name=$SERVER_NAME"
```

## [AI Admin] Create customer Registry on Llama Stack NS
Create the registry NS:
```
oc new-project mcp-demo-registry
```

And the `McpRegistry` in this namespace:
```
oc apply -n mcp-demo-registry -f demo/registry.yaml
```

**TODO**:Automate manual steps
* Grant permission to pull images from `mcp-catalog` registry
```
oc policy add-role-to-user system:image-puller system:serviceaccount:mcp-demo-registry:default -n mcp-catalog
```

## [AI Admin] Activate server for blueprint(s)
Add a secret holding the required `TAVILY_API_KEY` (you must set this in the environment):
```
oc create secret generic mcp-tavily-secret -n mcp-demo-registry --from-literal=TAVILY_API_KEY=$TAVILY_API_KEY
```
Finally, create the `McpServer` instance:
```
oc apply -n mcp-demo-registry -f demo/tavily-server.yaml
```

## [AI Engineer] Run MCP tools from Notebook
Run [demo.ipynb](demo.ipynb) Notebook.

## [AI Engineer] Register server(s) to Llama Stack
## [AI Engineer] Run agentic workflow from customer Notebook

