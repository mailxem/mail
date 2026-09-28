export { default as EmailFlowBuilder } from "@/components/emailFlowBuilder/EmailFlowBuilder";
export { default as TurboNode } from "@/components/emailFlowBuilder/nodes/TurboNode";
export { default as TurboEdge } from "@/components/emailFlowBuilder/edges/TurboEdge";
export { nodeTypes } from "@/components/emailFlowBuilder/nodes/NodeTypeSelector";
export type { TurboNodeData, NodeConfig, EmailFlowState } from "@/components/emailFlowBuilder/types";

// Re-export necessary React Flow types that consumers might need
export type {
  Node,
  Edge,
  Connection,
  NodeProps,
  EdgeProps,
} from '@xyflow/react'; 