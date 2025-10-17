import {Cell, Legend, Pie, PieChart, LabelList} from 'recharts';
import type {Sensor} from "../types/types.ts";

interface SensorHealthPieChartProps {
  sensors: Sensor[]
}

function SensorHealthPieChart({sensors}: SensorHealthPieChartProps) {

  const COLORS = ['#00C49F',  '#FF8042', '#FFBB28'];
  console.log(sensors);
  const data = [
    { name: 'Good', value: 0 },
    { name: 'Bad', value: 0 },
    { name: 'Unknown', value: 0 },
  ];

  for (const sensor of sensors) {
    if (sensor.healthStatus === 'good') {
      data[0].value += 1;
    } else if (sensor.healthStatus === 'bad') {
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
        <Legend verticalAlign="top" height={36}/>
        <LabelList dataKey="value" position="outside" />
        {data.map((entry, index) => (
          <Cell key={`cell-${entry.name}`} fill={COLORS[index % COLORS.length]} />
        ))}
      </Pie>
    </PieChart>
  );
}

export default SensorHealthPieChart;