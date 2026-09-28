import React, { useCallback, useState } from "react";
import {
  ReactFlow,
  Controls,
  Background,
  useNodesState,
  useEdgesState,
  addEdge,
  BackgroundVariant,
  type Node,
  type Edge,
  type Connection,
  type XYPosition,
  ConnectionMode,
  useReactFlow,
  NodeTypes,
} from "@xyflow/react";
import { FiPlusCircle } from "react-icons/fi";
import { IconType } from "react-icons";

import TurboNode from "@/components/emailFlowBuilder/nodes/TurboNode";
import TurboEdge from "@/components/emailFlowBuilder/edges/TurboEdge";
import NodeTypeSelector, {
  nodeTypes as availableNodeTypes,
} from "@/components/emailFlowBuilder/nodes/NodeTypeSelector";
import {
  initialNodes,
  initialEdges,
} from "@/components/emailFlowBuilder/utils/initial-elements";
import NodeContextMenu from "@/components/emailFlowBuilder/nodes/NodeContextMenu";
import NodeConfigForm from "@/components/emailFlowBuilder/nodes/NodeConfigForm";
import { TurboNodeData } from "@/components/emailFlowBuilder/types";

// Define type for node types
type NodeType = {
  icon: IconType;
  title: string;
  description: string;
  config?: any;
};

type NodeTypeMap = {
  [key: string]: NodeType;
};

const nodeTypes: NodeTypes = {
  turbo: TurboNode,
};

const edgeTypes = {
  turbo: TurboEdge,
};

const defaultEdgeOptions = {
  type: "turbo",
  animated: true,
};

const EmailFlowBuilder = () => {
  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);
  const [selectorOpen, setSelectorOpen] = useState(false);
  const [selectorPosition, setSelectorPosition] = useState<XYPosition>({
    x: 0,
    y: 0,
  });
  const [selectedNode, setSelectedNode] = useState<Node<TurboNodeData> | null>(
    null
  );
  const [contextMenu, setContextMenu] = useState<{
    node: Node<TurboNodeData>;
    position: XYPosition;
  } | null>(null);

  const onConnect = useCallback(
    (params: Connection) => setEdges((eds) => addEdge(params, eds)),
    [setEdges]
  );

  const onEdgeUpdate = useCallback(
    (oldEdge: Edge, newConnection: Connection) => {
      if (newConnection.source === newConnection.target) return;

      setEdges((els) => {
        const newEdges = els.filter((e) => e.id !== oldEdge.id);
        return addEdge(
          {
            ...newConnection,
            id: `e${newConnection.source}-${newConnection.target}`,
            animated: true,
          },
          newEdges
        );
      });
    },
    [setEdges]
  );

  const [zoom, setZoom] = useState(1);

  const screenToFlowPosition = useCallback((position: XYPosition) => {
    const { x, y } = position;
    return {
      x: x / zoom,
      y: y / zoom,
    };
  }, []);

  const addNode = useCallback(
    (type: string) => {
      const nodeTypeInfo = (availableNodeTypes as NodeTypeMap)[type];
      if (!nodeTypeInfo) return;

      const position = screenToFlowPosition({
        x: selectorPosition.x,
        y: selectorPosition.y,
      });

      const newNode: Node<TurboNodeData> = {
        id: `${nodes.length + 1}`,
        type: "turbo",
        position,
        data: {
          icon: React.createElement(nodeTypeInfo.icon),
          title: nodeTypeInfo.title,
          subline: "Configure step",
          type: type,
          description: nodeTypeInfo.description,
          config: nodeTypeInfo.config,
        },
      };

      setNodes((nds) => [...nds, newNode]);

      if (selectedNode) {
        const edge: Edge = {
          id: `e${selectedNode.id}-${newNode.id}`,
          source: selectedNode.id,
          target: newNode.id,
          animated: true,
        };
        setEdges((eds) => addEdge(edge, eds));
      }
    },
    [
      nodes,
      selectorPosition,
      selectedNode,
      screenToFlowPosition,
      setNodes,
      setEdges,
    ]
  );

  const calculatePopupPosition = useCallback((buttonRect: DOMRect) => {
    const POPUP_WIDTH = 300;
    const POPUP_HEIGHT = 400;
    const PADDING = 20;

    const viewportWidth = window.innerWidth;
    const viewportHeight = window.innerHeight;

    let x = buttonRect.x;
    let y = buttonRect.y - POPUP_HEIGHT - PADDING;

    if (x + POPUP_WIDTH > viewportWidth - PADDING) {
      x = viewportWidth - POPUP_WIDTH - PADDING;
    }
    if (x < PADDING) {
      x = PADDING;
    }

    if (y < PADDING) {
      y = buttonRect.bottom + PADDING;
    }

    return { x, y };
  }, []);

  const onNodeContextMenu = useCallback(
    (event: React.MouseEvent, node: Node) => {
      event.preventDefault();
      const bounds = event.currentTarget.getBoundingClientRect();
      setContextMenu({
        node: node as Node<TurboNodeData>,
        position: { x: bounds.right, y: bounds.top },
      });
    },
    []
  );

  const deleteNode = useCallback(
    (nodeId: string) => {
      setNodes((nds) => nds.filter((node) => node.id !== nodeId));
      setEdges((eds) =>
        eds.filter((edge) => edge.source !== nodeId && edge.target !== nodeId)
      );
    },
    [setNodes, setEdges]
  );

  const updateNode = useCallback(
    (nodeId: string, newData: Partial<TurboNodeData>) => {
      setNodes((nds) =>
        nds.map((node) => {
          if (node.id === nodeId) {
            return {
              ...node,
              data: {
                ...node.data,
                ...newData,
              },
            };
          }
          return node;
        })
      );
    },
    [setNodes]
  );

  return (
    <div style={{ width: "100%", height: "100%" }}>
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onConnect={onConnect}
        onNodeClick={(_, node) => setSelectedNode(node as Node<TurboNodeData>)}
        onNodeContextMenu={onNodeContextMenu}
        onPaneClick={() => {
          setSelectedNode(null);
          setContextMenu(null);
        }}
        fitView
        nodeTypes={nodeTypes}
        edgeTypes={edgeTypes}
        defaultEdgeOptions={defaultEdgeOptions}
        connectionMode={ConnectionMode.Loose}
        defaultViewport={{ x: 0, y: 0, zoom: 1 }}
        snapToGrid={true}
        snapGrid={[15, 15]}
        style={{ backgroundColor: "#F7F9FB" }}
      >
        <Controls showInteractive={true} />
        <Background variant={BackgroundVariant.Dots} gap={12} size={0.6} />
      </ReactFlow>

      <button
        className="add-node-button"
        onClick={(e) => {
          const rect = e.currentTarget.getBoundingClientRect();
          setSelectorPosition(calculatePopupPosition(rect));
          setSelectorOpen(true);
        }}
      >
        <FiPlusCircle />
        <span>Add Campaign Step</span>
      </button>

      <NodeTypeSelector
        isOpen={selectorOpen}
        onClose={() => setSelectorOpen(false)}
        onSelect={addNode}
        position={selectorPosition}
      />

      {contextMenu && (
        <NodeContextMenu
          position={contextMenu.position}
          onEdit={() => {
            setSelectedNode(contextMenu.node);
            setContextMenu(null);
          }}
          onDelete={() => {
            deleteNode(contextMenu.node.id);
            setContextMenu(null);
          }}
          onClose={() => setContextMenu(null)}
        />
      )}

      {selectedNode && (
        <NodeConfigForm
          node={selectedNode}
          onUpdate={updateNode}
          onClose={() => setSelectedNode(null)}
        />
      )}
    </div>
  );
};

export default EmailFlowBuilder;
