import { NavLink, Outlet } from "react-router";
import { House, MapPin, Package, Search, Tags } from "lucide-react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useTheme } from "@/lib/use-theme";

const navItems = [
  { to: "/", label: "Home", icon: House },
  { to: "/locations", label: "Locations", icon: MapPin },
  { to: "/items", label: "Items", icon: Package },
  { to: "/search", label: "Search", icon: Search },
  { to: "/tags", label: "Tags", icon: Tags },
];

export function Layout() {
  const { theme, setTheme } = useTheme();

  return (
    <div className="min-h-screen bg-[radial-gradient(circle_at_top,oklch(0.98_0.01_240)_0%,var(--background)_38%)]">
      <header className="sticky top-0 z-50 border-b bg-background/90 shadow-sm backdrop-blur supports-[backdrop-filter]:bg-background/70">
        <div className="mx-auto flex h-16 max-w-6xl items-center gap-6 px-4">
          <NavLink to="/" className="flex items-center gap-2.5 font-semibold">
            <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground shadow-sm">
              <Package className="h-4 w-4" />
            </span>
            <span className="tracking-tight">Cubby</span>
          </NavLink>

          <nav className="flex items-center gap-1 overflow-x-auto">
            {navItems.map(({ to, label, icon: Icon }) => (
              <NavLink
                key={to}
                to={to}
                end={to === "/"}
                className={({ isActive }) =>
                  `flex shrink-0 items-center gap-1.5 whitespace-nowrap rounded-full px-3 py-1.5 text-sm transition-all ${
                    isActive
                      ? "bg-primary text-primary-foreground shadow-sm"
                      : "text-muted-foreground hover:bg-accent/60 hover:text-foreground"
                  }`
                }
              >
                <Icon className="h-4 w-4" />
                {label}
              </NavLink>
            ))}
          </nav>

          <div className="ml-auto">
            <Select
              value={theme}
              onValueChange={(value) => setTheme(value === "dark" ? "dark" : "light")}
            >
              <SelectTrigger className="h-9 w-[120px]">
                <SelectValue placeholder="Theme" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="light">Light</SelectItem>
                <SelectItem value="dark">Dark</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
      </header>

      <main className="mx-auto max-w-6xl px-4 py-6 md:py-8">
        <div className="rounded-2xl border bg-card/80 p-5 shadow-sm backdrop-blur-sm md:p-6">
          <Outlet />
        </div>
      </main>
    </div>
  );
}
