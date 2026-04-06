import { Cell, Legend, Pie, PieChart, LabelList, ResponsiveContainer } from "recharts";
import type { Sensor } from "../types/types.ts";
import { useChartColours } from "../theme/chartColours";

interface SensorTypePieChartProps {
  sensors: Sensor[];
}

function SensorTypePieChart({ sensors }: SensorTypePieChartProps) {
  const chartColours = useChartColours();
  const COLORS = chartColours.categorical;

  const data = [
    { name: "sensor-hub-http-temperature", value: 0 },
    { name: "Other", value: 0 },
  ];

  for (const sensor of sensors) {
    if (sensor.sensorDriver === data[0].name) {
      data[0].value += 1;
    } else {
      data[1].value += 1;
    }
  }

  return (
    <ResponsiveContainer width="100%" height="100%">
      <PieChart>
        <Pie
          data={data}
          innerRadius="40%"
          outerRadius="55%"
          fill="#8884d8"
          paddingAngle={5}
          dataKey="value"
          nameKey="name"
          cx="50%"
          cy="50%"
        >
          <Legend verticalAlign="top" height={36} />
          <LabelList dataKey="value" position="outside" />
          {data.map((entry, index) => (
            <Cell
              key={`cell-${entry.name}`}
              fill={COLORS[index % COLORS.length]}
            />
          ))}
        </Pie>
      </PieChart>
    </ResponsiveContainer>
  );
}

export default SensorTypePieChart;
