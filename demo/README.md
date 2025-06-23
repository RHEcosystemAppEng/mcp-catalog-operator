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
* Install RBAC resources (default):
```
oc apply -k config/rbac
```
* Install of build pipeline:
```
oc apply -f demo/pipeline.yaml -n mcp-catalog
```
* Bind `privileged` SCC to `pipeline` service account:
```
oc adm policy add-scc-to-user privileged \
  -z pipeline -n mcp-catalog
```

## [Admin] Create global Catalog

Install a sample catalog instance:
```
oc apply -n mcp-catalog -f demo/catalog.yaml
```

## [Admin] Ingest server defs from reference MCP registry
Run the import service from the catalog service:
```
MCP_REGISTRY=$(oc get route -n mcp-registry-ref mcp-registry-route -oyaml | yq '.status.ingress[0].host') && \
  echo "https://$MCP_REGISTRY/v0" && \
  curl -X POST "localhost:8000/import?mcp_registry_source=https://$MCP_REGISTRY/v0" | jq
```

Verify the CRs:
```
oc get mcpserverdefinitions -n mcp-catalog | wc -l
```

## [Admin] Promote selected instance(s) to blueprint
Expose the registry service port:
```
oc port-forward svc/red-hat-ecosystem-mcp-catalog-svc 8000:8000
```

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
oc adm policy add-role-to-user system:image-puller system:serviceaccount:mcp-demo-registry:default -n mcp-catalog
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
**NOTE** Remember to update the blueprint ID!

**TODO**:
* Use `MCP Catalog` service to create the instance.

## [AI Engineer] Run MCP tools from Notebook
Run [demo.ipynb](demo.ipynb) Notebook.

## [AI Engineer] Register server(s) to Llama Stack

```
ollama serve
```

```
ollama run llama3.2:3b-instruct-fp16 --keepalive 240m
```

```
# Enable ollama and MCP providers
llama stack build
```

```
export INFERENCE_MODEL=llama3.2:3b-instruct-fp16
llama stack run /Users/dmartino/.llama/distributions/llamastack-mcp-demo-stack/llamastack-mcp-demo-stack-run.yaml
```

## [AI Engineer] Run agentic workflow from customer Notebook
Run [lls-mcp-demo.ipynb](./lls-mcp-demo.ipynb) Notebook.


