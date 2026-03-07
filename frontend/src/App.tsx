import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { BrowserRouter, Routes, Route } from "react-router";
import { Layout } from "@/components/layout";
import { HomePage } from "@/pages/home";
import { LocationsPage } from "@/pages/locations";
import { ItemsPage } from "@/pages/items";
import { ItemDetailPage } from "@/pages/item-detail";
import { SearchPage } from "@/pages/search";
import { TagsPage } from "@/pages/tags";
import { ThemeProvider } from "@/lib/theme-provider";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      retry: 1,
    },
  },
});

function App() {
  return (
    <ThemeProvider>
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>
          <Routes>
            <Route element={<Layout />}>
              <Route index element={<HomePage />} />
              <Route path="home" element={<HomePage />} />
              <Route path="locations" element={<LocationsPage />} />
              <Route path="items" element={<ItemsPage />} />
              <Route path="items/:id" element={<ItemDetailPage />} />
              <Route path="search" element={<SearchPage />} />
              <Route path="tags" element={<TagsPage />} />
            </Route>
          </Routes>
        </BrowserRouter>
      </QueryClientProvider>
    </ThemeProvider>
  );
}

export default App;
