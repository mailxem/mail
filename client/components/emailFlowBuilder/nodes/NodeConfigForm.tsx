import React from "react";
import { Node } from "@xyflow/react";
import { TurboNodeData } from "@/components/emailFlowBuilder/types";
import {
  DropdownMenu,
  DropdownMenuItem,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from "@radix-ui/react-dropdown-menu";
import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";

interface NodeConfigFormProps {
  node: Node<TurboNodeData>;
  onUpdate: (nodeId: string, data: Partial<TurboNodeData>) => void;
  onClose: () => void;
}

interface DelayConfig {
  duration: number;
  unit: "minutes" | "hours" | "days" | "weeks";
}

interface EmailConfig {
  subject: string;
  template: string;
}

interface ConditionConfig {
  type: "email_opened" | "email_clicked" | "custom";
  value?: string;
  operator?: string;
}

interface FilterConfig {
  criteria: string;
  value: string;
  operator: "equals" | "contains" | "greater_than" | "less_than";
}

interface TagConfig {
  action: "add" | "remove";
  tags: string[];
}

interface AiConfig {
  prompt: string;
  model: string;
  temperature: number;
  maxTokens: number;
}

type NodeConfig =
  | DelayConfig
  | EmailConfig
  | ConditionConfig
  | FilterConfig
  | TagConfig
  | AiConfig;

const NodeConfigForm: React.FC<NodeConfigFormProps> = ({
  node,
  onUpdate,
  onClose,
}) => {
  const getDefaultConfig = (type: string): NodeConfig => {
    switch (type) {
      case "delay":
        return { duration: 1, unit: "days" };
      case "email":
        return { subject: "", template: "default" };
      case "condition":
        return { type: "email_opened" };
      case "filter":
        return { criteria: "", value: "", operator: "equals" };
      case "tag":
        return { action: "add", tags: [] };
      case "ai":
        return {
          prompt: "",
          model: "gpt-3.5-turbo",
          temperature: 0.7,
          maxTokens: 500,
        };
      default:
        return { duration: 1, unit: "days" }; // Default to delay config
    }
  };

  const [config, setConfig] = React.useState<NodeConfig>(() => {
    const defaultConfig = getDefaultConfig(node.data.type || "delay");
    return node.data.config ? defaultConfig : defaultConfig;
  });

  const renderDelayConfig = () => {
    const delayConfig = config as DelayConfig;
    return (
      <>
        <label>Duration</label>
        <div className="input-group">
          <input
            type="number"
            min="1"
            value={delayConfig.duration}
            onChange={(e) =>
              setConfig({ ...delayConfig, duration: parseInt(e.target.value) })
            }
          />
          <select
            value={delayConfig.unit}
            onChange={(e) =>
              setConfig({
                ...delayConfig,
                unit: e.target.value as DelayConfig["unit"],
              })
            }
          >
            <option value="minutes">Minutes</option>
            <option value="hours">Hours</option>
            <option value="days">Days</option>
            <option value="weeks">Weeks</option>
          </select>
        </div>
      </>
    );
  };

  const renderEmailConfig = () => {
    const emailConfig = config as EmailConfig;
    return (
      <>
        <label>Email Subject</label>
        <input
          type="text"
          value={emailConfig.subject}
          onChange={(e) =>
            setConfig({ ...emailConfig, subject: e.target.value })
          }
          placeholder="Enter email subject"
        />
        <label>Template</label>
        <select
          value={emailConfig.template}
          onChange={(e) =>
            setConfig({ ...emailConfig, template: e.target.value })
          }
        >
          <option value="default">Default Template</option>
          <option value="promotional">Promotional</option>
          <option value="newsletter">Newsletter</option>
          <option value="announcement">Announcement</option>
        </select>
      </>
    );
  };

  const renderConditionConfig = () => {
    const conditionConfig = config as ConditionConfig;
    return (
      <>
        <label>Condition Type</label>
        <select
          value={conditionConfig.type}
          onChange={(e) =>
            setConfig({
              ...conditionConfig,
              type: e.target.value as ConditionConfig["type"],
            })
          }
        >
          <option value="email_opened">Email Opened</option>
          <option value="email_clicked">Email Clicked</option>
          <option value="custom">Custom</option>
        </select>
        {conditionConfig.type === "custom" && (
          <>
            <label>Custom Condition</label>
            <input
              type="text"
              value={conditionConfig.value || ""}
              onChange={(e) =>
                setConfig({ ...conditionConfig, value: e.target.value })
              }
              placeholder="Enter custom condition"
            />
          </>
        )}
      </>
    );
  };

  const renderFilterConfig = () => {
    const filterConfig = config as FilterConfig;
    return (
      <>
        <label>Filter Criteria</label>
        <input
          type="text"
          value={filterConfig.criteria}
          onChange={(e) =>
            setConfig({ ...filterConfig, criteria: e.target.value })
          }
          placeholder="e.g., user.country"
        />
        <label>Operator</label>
        <select
          value={filterConfig.operator}
          onChange={(e) =>
            setConfig({
              ...filterConfig,
              operator: e.target.value as FilterConfig["operator"],
            })
          }
        >
          <option value="equals">Equals</option>
          <option value="contains">Contains</option>
          <option value="greater_than">Greater Than</option>
          <option value="less_than">Less Than</option>
        </select>
        <label>Value</label>
        <input
          type="text"
          value={filterConfig.value}
          onChange={(e) =>
            setConfig({ ...filterConfig, value: e.target.value })
          }
          placeholder="Enter value"
        />
      </>
    );
  };

  const renderTagConfig = () => {
    const tagConfig = config as TagConfig;
    return (
      <>
        <label>Action</label>
        <select
          value={tagConfig.action}
          onChange={(e) =>
            setConfig({
              ...tagConfig,
              action: e.target.value as TagConfig["action"],
            })
          }
        >
          <option value="add">Add Tags</option>
          <option value="remove">Remove Tags</option>
        </select>
        <label>Tags</label>
        <input
          type="text"
          value={tagConfig.tags.join(", ")}
          onChange={(e) =>
            setConfig({
              ...tagConfig,
              tags: e.target.value.split(",").map((t) => t.trim()),
            })
          }
          placeholder="Enter tags, comma separated"
        />
      </>
    );
  };

  const renderConfig = () => {
    switch (node.data.type) {
      case "delay":
        return renderDelayConfig();
      case "email":
        return renderEmailConfig();
      case "condition":
        return renderConditionConfig();
      case "filter":
        return renderFilterConfig();
      case "tag":
        return renderTagConfig();
      case "ai":
        return renderAiConfig();
      default:
        return null;
    }
  };

  const renderAiConfig = () => {
    const aiConfig = config as AiConfig;
    return (
      <>
        <label>Prompt</label>
        <Textarea
          value={aiConfig.prompt}
          onChange={(e) => setConfig({ ...aiConfig, prompt: e.target.value })}
          placeholder="Enter prompt"
        />
        <label>Model</label>
        <Select
          value={aiConfig.model}
          onValueChange={(value) => setConfig({ ...aiConfig, model: value })}
        >
          <SelectTrigger>
            <SelectValue placeholder="Select model" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="gpt-3.5-turbo">GPT-3.5 Turbo</SelectItem>
            <SelectItem value="gpt-4">GPT-4</SelectItem>
            <SelectItem value="gpt-4o">GPT-4o</SelectItem>
          </SelectContent>
        </Select>
      </>
    );
  };

  const getConfigSummary = (type: string, nodeConfig: NodeConfig): string => {
    switch (type) {
      case "delay": {
        const delayConfig = nodeConfig as DelayConfig;
        return `Wait ${delayConfig.duration} ${delayConfig.unit}`;
      }
      case "email": {
        const emailConfig = nodeConfig as EmailConfig;
        return `Subject: ${emailConfig.subject}`;
      }
      case "condition": {
        const conditionConfig = nodeConfig as ConditionConfig;
        return conditionConfig.type === "custom"
          ? conditionConfig.value || ""
          : conditionConfig.type;
      }
      case "filter": {
        const filterConfig = nodeConfig as FilterConfig;
        return `${filterConfig.criteria} ${filterConfig.operator} ${filterConfig.value}`;
      }
      case "tag": {
        const tagConfig = nodeConfig as TagConfig;
        return `${tagConfig.action} ${tagConfig.tags.join(", ")}`;
      }
      case "ai": {
        const aiConfig = nodeConfig as AiConfig;
        return `${aiConfig.prompt} (${aiConfig.model})`;
      }
      default:
        return "";
    }
  };

  const handleSubmit = () => {
    onUpdate(node.id, {
      ...node.data,
      config,
      subline: getConfigSummary(node.data.type || "delay", config),
    });
    onClose();
  };

  return (
    <div className="node-edit-form">
      <h3>{node.data.title} Configuration</h3>
      <div className="node-config-form">
        <label>Title</label>
        <input
          type="text"
          value={node.data.title}
          onChange={(e) => onUpdate(node.id, { title: e.target.value })}
        />
        {renderConfig()}
      </div>
      <div className="form-buttons">
        <button onClick={handleSubmit} className="primary">
          Save
        </button>
        <button onClick={onClose} className="secondary">
          Cancel
        </button>
      </div>
    </div>
  );
};

export default NodeConfigForm;
