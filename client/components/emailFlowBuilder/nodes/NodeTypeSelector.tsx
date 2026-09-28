import React from "react";
import {
  FiMail,
  FiClock,
  FiUsers,
  FiCheck,
  FiGitBranch,
  FiSend,
  FiMessageSquare,
  FiFilter,
  FiTag,
  FiCpu,
} from "react-icons/fi";
import { NodeTypeSelectorProps } from '@/components/emailFlowBuilder/types';

export const nodeTypes = {
  ai: {
    icon: FiCpu,
    title: "AI",
    description: "Create and send email content",
    config: {
      prompt: "",
      model: "gpt-3.5-turbo",
      temperature: 0.7,
      maxTokens: 500,
    },
  },
  audience: {
    icon: FiUsers,
    title: "Segment",
    description: "Define your target audience",
  },
  email: {
    icon: FiMail,
    title: "Email Campaign",
    description: "Create and send email content",
  },
  delay: {
    icon: FiClock,
    title: "Time Delay",
    description: "Add waiting period between steps",
  },
  condition: {
    icon: FiGitBranch,
    title: "Split Test",
    description: "Create A/B tests or conditions",
  },
  action: {
    icon: FiSend,
    title: "Trigger",
    description: "Start automated actions",
  },
  notification: {
    icon: FiMessageSquare,
    title: "Alert",
    description: "Send team notifications",
  },
  filter: {
    icon: FiFilter,
    title: "Filter",
    description: "Refine audience based on behavior",
  },
  tag: {
    icon: FiTag,
    title: "Tag Manager",
    description: "Add/remove subscriber tags",
  },
  end: {
    icon: FiCheck,
    title: "Goal",
    description: "Track campaign completion",
  },
};

const NodeTypeSelector: React.FC<NodeTypeSelectorProps> = ({
  isOpen,
  onClose,
  onSelect,
  position,
}) => {
  if (!isOpen) return null;

  return (
    <div
      className="node-type-selector"
      style={{
        position: "absolute",
        left: position.x,
        top: position.y - 10,
        bottom: 10,
      }}
    >
      <div className="node-type-selector-header">
        <h3>Add Campaign Step</h3>
        <button onClick={onClose} className="close-button">
          ×
        </button>
      </div>
      <div className="node-type-selector-content">
        {Object.entries(nodeTypes).map(([type, info]) => (
          <button
            key={type}
            className="node-type-button"
            onClick={() => {
              onSelect(type);
              onClose();
            }}
          >
            <span className="node-type-icon">
              {React.createElement(info.icon)}
            </span>
            <div className="node-type-info">
              <span className="node-type-title">{info.title}</span>
              <span className="node-type-description">{info.description}</span>
            </div>
          </button>
        ))}
      </div>
    </div>
  );
};

export default NodeTypeSelector; 