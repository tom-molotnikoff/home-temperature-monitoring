import { Cell, Legend, Pie, PieChart, LabelList } from "recharts";
import type { Sensor } from "../types/types.ts";

interface SensorTypePieChartProps {
  sensors: Sensor[];
}

function SensorTypePieChart({ sensors }: SensorTypePieChartProps) {
  const COLORS = ["#00c458ff", "#cecb00ff", "#28cdffff"];

  const data = [
    { name: "Temperature", value: 0 },
    { name: "Humidity", value: 0 },
    { name: "Other", value: 0 },
  ];

  for (const sensor of sensors) {
    if (sensor.type === data[0].name) {
      data[0].value += 1;
    } else if (sensor.type === data[1].name) {
      data[1].value += 1;
    } else {
      data[2].value += 1;
    }
  }

  return (
    <PieChart width={300} height={250}>
      <Pie
        data={data}
        innerRadius={60}
        outerRadius={80}
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
  );
}

export default SensorTypePieChart;
