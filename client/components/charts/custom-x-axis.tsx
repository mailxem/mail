"use client";

import { XAxis } from "recharts";

export function CustomXAxis() {
  return (
    <XAxis 
      dataKey="name" 
      stroke="#888888" 
      fontSize={12} 
      tickLine={false} 
      axisLine={false}
      padding={{ left: 10, right: 10 }} 
    />
  );
}