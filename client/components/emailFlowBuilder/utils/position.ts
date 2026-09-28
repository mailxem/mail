export const calculateFlowPosition = (
  event: MouseEvent,
  reactFlowBounds: DOMRect | undefined,
  zoom: number = 1
) => {
  if (!reactFlowBounds) return { x: 0, y: 0 };

  const position = {
    x: (event.clientX - reactFlowBounds.left) / zoom,
    y: (event.clientY - reactFlowBounds.top) / zoom
  };

  return position;
}; 