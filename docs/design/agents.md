# Agents

``` mermaid
stateDiagram-v2
    [*] --> Registered : First Initialization
    Registered --> Ready : Agent API Call
    Ready --> NotReady : Agent Operator Timeout
    NotReady --> Ready : Agent API Call

```
