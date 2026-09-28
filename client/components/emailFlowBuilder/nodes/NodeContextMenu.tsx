import React from 'react';
import { FiEdit2, FiTrash2, FiLink } from 'react-icons/fi';

interface NodeContextMenuProps {
  position: { x: number; y: number };
  onEdit: () => void;
  onDelete: () => void;
  onClose: () => void;
}

const NodeContextMenu: React.FC<NodeContextMenuProps> = ({
  position,
  onEdit,
  onDelete,
  onClose,
}) => {
  React.useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      const target = e.target as HTMLElement;
      if (!target.closest('.node-context-menu')) {
        onClose();
      }
    };

    document.addEventListener('click', handleClickOutside);
    return () => document.removeEventListener('click', handleClickOutside);
  }, [onClose]);

  return (
    <div
      className="node-context-menu"
      style={{
        position: 'absolute',
        left: position.x,
        top: position.y,
      }}
    >
      <button onClick={onEdit} className="context-menu-button">
        <FiEdit2 />
        <span>Edit Node</span>
      </button>
      <button onClick={onDelete} className="context-menu-button">
        <FiTrash2 />
        <span>Delete Node</span>
      </button>
    </div>
  );
};

export default NodeContextMenu; 