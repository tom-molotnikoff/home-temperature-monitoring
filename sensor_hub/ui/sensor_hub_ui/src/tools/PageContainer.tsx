import {type CSSProperties} from "react";
import { Box, CircularProgress } from "@mui/material";
import LayoutCard from "./LayoutCard.tsx";
import NavigationSidebar from "../navigation/NavigationSidebar.tsx";
import TopAppBar from "../navigation/TopAppBar.tsx";

interface PageContainerProps {
  children: React.ReactNode;
  titleText: string;
  loading?: boolean;
}

function PageContainer({ children, titleText, loading = false }: PageContainerProps) {
  return (
    <>
      <TopAppBar pageTitle={titleText} />
      <NavigationSidebar />
      <LayoutCard changes={layoutCardStyleChanges}>
        {loading ? (
          <Box sx={{ display: "flex", justifyContent: "center", alignItems: "center", flexGrow: 1, p: 4 }}>
            <CircularProgress />
          </Box>
        ) : (
          children
        )}
      </LayoutCard>
    </>
  );
}

const layoutCardStyleChanges: CSSProperties = {
  padding: "20px",
  alignItems: "stretch",
  minHeight: "calc(100vh - 65px)",
  border: "none",
  borderRadius: 0,
  width: "100%",
};

export default PageContainer;
