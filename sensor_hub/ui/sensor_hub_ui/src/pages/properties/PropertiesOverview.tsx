import PageContainer from '../../tools/PageContainer';
import { useAuth } from '../../providers/AuthContext';
import PropertiesCard from '../../components/PropertiesCard';

export default function PropertiesOverview() {
  const { user } = useAuth();

  return (
    <PageContainer titleText="Properties Overview" loading={user === undefined}>
      <PropertiesCard />
    </PageContainer>
  );
}
