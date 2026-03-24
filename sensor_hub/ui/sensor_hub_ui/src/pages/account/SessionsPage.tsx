import PageContainer from '../../tools/PageContainer';
import SessionsCard from '../../components/SessionsCard';

export default function SessionsPage() {
  return (
    <PageContainer titleText="Active sessions">
      <SessionsCard />
    </PageContainer>
  );
}
