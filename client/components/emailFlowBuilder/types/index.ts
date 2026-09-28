import { ReactNode } from 'react';
import { Node, Edge } from '@xyflow/react';

export interface TurboNodeData extends Record<string, unknown> {
  title: string;
  icon?: ReactNode;
  subline?: string;
  type?: string;
  description?: string;
  config?: NodeConfig;
}

export interface NodeConfig {
  duration?: number;
  unit?: 'minutes' | 'hours' | 'days' | 'weeks';
  subject?: string;
  template?: string;
  type?: 'email_opened' | 'email_clicked' | 'custom';
  value?: string;
  operator?: string;
  criteria?: string;
  action?: 'add' | 'remove';
  tags?: string[];
  audienceSize?: number;
  segmentName?: string;
  prompt?: string;
  model?: string;
  temperature?: number;
  maxTokens?: number;
}

export interface EmailFlowState {
  nodes: Node<TurboNodeData>[];
  edges: Edge[];
}

export interface NodeTypeSelectorProps {
  isOpen: boolean;
  onClose: () => void;
  onSelect: (type: string) => void;
  position: { x: number; y: number };
} 