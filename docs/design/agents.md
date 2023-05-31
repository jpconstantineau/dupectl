# Agents

``` mermaid
stateDiagram-v2
    [*] --> Created : First Initialization
    Created --> Ready : Agent API Call
    Ready --> NotReady : Agent Operator Timeout
    NotReady --> Ready : Agent API Call

```
