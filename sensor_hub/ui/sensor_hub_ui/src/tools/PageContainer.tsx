import {type CSSProperties} from "react";
import LayoutCard from "./LayoutCard.tsx";
import NavigationSidebar from "../navigation/NavigationSidebar.tsx";
import TopAppBar from "../navigation/TopAppBar.tsx";

interface PageContainerProps {
  children: React.ReactNode;
  titleText: string;
}

function PageContainer({ children, titleText }: PageContainerProps) {

  return (
    <>
      <TopAppBar pageTitle={titleText}/>
      <NavigationSidebar/>
      <LayoutCard changes={layoutCardStyleChanges}>
        {children}
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
