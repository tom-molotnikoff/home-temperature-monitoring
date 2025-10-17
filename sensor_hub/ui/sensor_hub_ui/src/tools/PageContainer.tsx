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
  alignItems: "center",
  border: "none",
  borderRadius: 0,
};

export default PageContainer;
