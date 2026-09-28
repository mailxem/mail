import React from 'react';
import { Node, Edge } from '@xyflow/react';
import { FiUsers, FiMail, FiClock, FiGitBranch, FiCheck, FiCpu } from 'react-icons/fi';
import { TurboNodeData } from '@/components/emailFlowBuilder/types';

const createIcon = (Icon: React.ComponentType) => React.createElement(Icon);

export const initialNodes: Node<TurboNodeData>[] = [
  {
    id: "1",
    position: { x: 0, y: 150 },
    data: {
      icon: createIcon(FiUsers),
      title: "New Subscribers", 
      subline: "Entry Point",
      type: "audience",
      description: "New users in last 30 days",
      config: {
        audienceSize: 1000,
        segmentName: "New Subscribers",
      },
    },
    type: "turbo",
  },
  {
    id: "2", 
    position: { x: 300, y: 0 },
    data: {
      icon: createIcon(FiMail),
      title: "Welcome Series",
      subline: "First Email",
      type: "email",
      description: "Initial welcome campaign",
      config: {
        subject: "Welcome to Our Community!",
        template: "welcome_template",
      },
    },
    type: "turbo",
  },
  {
    id: "3",
    position: { x: 300, y: 250 },
    data: {
      icon: createIcon(FiClock),
      title: "Engagement Wait",
      subline: "3 Day Delay", 
      type: "delay",
      description: "Wait for email interaction",
      config: { duration: 3, unit: "days" },
    },
    type: "turbo",
  },
  {
    id: "4",
    position: { x: 600, y: 150 },
    data: {
      icon: createIcon(FiCpu),
      title: "AI Email Generator",
      subline: "Generate Email Content",
      type: "ai",
      description: "Generate personalized email content using AI",
      config: {
        prompt: "Generate engaging email content",
        model: "gpt-3.5-turbo",
        temperature: 0.7,
        maxTokens: 500
      }
    },
    type: "turbo"
  },
  {
    id: "5",
    position: { x: 900, y: 300 },
    data: {
      icon: createIcon(FiGitBranch),
      title: "Engagement Check",
      subline: "Email Opened?",
      type: "condition",
      description: "Check email engagement",
      config: {
        type: "email_opened",
        value: "welcome_email",
      },
    },
    type: "turbo",
  },
  {
    id: "6",
    position: { x: 900, y: 0 },
    data: {
      icon: createIcon(FiMail),
      title: "Product Showcase",
      subline: "Engaged Users",
      type: "email",
      config: {
        subject: "Discover Our Best Features",
        template: "product_showcase",
      },
    },
    type: "turbo",
  },
  {
    id: "7",
    position: { x: 900, y: 300 },
    data: {
      icon: createIcon(FiMail),
      title: "Re-engagement",
      subline: "Recovery Email",
      type: "email",
      config: {
        subject: "Don't Miss Out!",
        template: "reengagement",
      },
    },
    type: "turbo",
  },
  {
    id: "8",
    position: { x: 1200, y: 150 },
    data: {
      icon: createIcon(FiCheck),
      title: "Campaign Complete",
      subline: "Journey End",
      type: "end",
      description: "Welcome series completed",
      config: {},
    },
    type: "turbo",
  },
];

export const initialEdges: Edge[] = [
  {
    id: "e1-2",
    source: "1",
    target: "2",
    animated: true,
  },
  {
    id: "e1-3",
    source: "1",
    target: "3",
    animated: true,
  },
  {
    id: "e2-4",
    source: "2",
    target: "4",
    animated: true,
  },
  {
    id: "e3-4",
    source: "3",
    target: "4",
    animated: true,
  },
  {
    id: "e4-5",
    source: "4",
    target: "5",
    label: "Opened",
    animated: true,
  },
  {
    id: "e4-6",
    source: "4",
    target: "6",
    label: "Not Opened",
    animated: true,
  },
  {
    id: "e5-7",
    source: "5",
    target: "7",
    animated: true,
  },
  {
    id: "e6-7",
    source: "6",
    target: "7",
    animated: true,
  },
  {
    id: "e7-8",
    source: "7",
    target: "8",
    animated: true,
  },
];
