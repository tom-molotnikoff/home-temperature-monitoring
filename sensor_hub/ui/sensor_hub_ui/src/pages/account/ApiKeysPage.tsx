import PageContainer from '../../tools/PageContainer';
import ApiKeysCard from '../../components/ApiKeysCard';
import { useApiKeys } from '../../hooks/useApiKeys';

export default function ApiKeysPage() {
  const { apiKeys, loaded, refresh } = useApiKeys();

  return (
    <PageContainer titleText="API Keys">
      <ApiKeysCard apiKeys={apiKeys} loaded={loaded} onRefresh={refresh} />
    </PageContainer>
  );
}
