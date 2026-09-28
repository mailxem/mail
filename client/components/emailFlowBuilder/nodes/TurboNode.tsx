import React, { memo } from 'react';
import { Handle, Position, type NodeProps } from '@xyflow/react';
import { TurboNodeData } from '@/components/emailFlowBuilder/types';

const TurboNode = ({ data }: NodeProps) => {
  const nodeData = data as TurboNodeData;
  
  return (
    <>
      <div className="cloud gradient">
        <div>{nodeData.icon}</div>
      </div>
      <div className="wrapper gradient">
        <div className="inner">
          <div className="body">
            {nodeData.icon && <div className="icon">{nodeData.icon}</div>}
            <div>
              <div className="title">{nodeData.title}</div>
              {nodeData.subline && <div className="subline">{nodeData.subline}</div>}
            </div>
          </div>
          <Handle type="target" position={Position.Left} />
          <Handle type="source" position={Position.Right} />
        </div>
      </div>
    </>
  );
};

export default memo(TurboNode); 