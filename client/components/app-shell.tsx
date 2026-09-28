"use client";

import { ReactNode, useEffect, useRef, useState } from "react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { signOut, useSession } from "next-auth/react";
import { useTheme } from "next-themes";
import {
  ArrowUpRight,
  CreditCard,
  KeyRound,
  Mailbox,
  Server,
  Tags,
  UserRound,
  Webhook,
  ChartNoAxesCombined,
  ChevronDown,
  ChevronsUpDown,
  CircleHelp,
  FilePenLine,
  LayoutGrid,
  LogOut,
  Mail,
  Menu,
  Monitor,
  Moon,
  Newspaper,
  PanelLeftClose,
  PanelLeftOpen,
  PanelsTopLeft,
  Search,
  Send,
  Settings,
  Sparkles,
  Sun,
  Users,
  Workflow,
} from "lucide-react";
import { AppHeader } from "@/components/app-header";
import { Modal } from "@/components/marketing/shared";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuSeparator,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetTitle,
} from "@/components/ui/sheet";
import { cn } from "@/lib/utils";
import styles from "@/components/workspace-shell.module.css";

const navigationGroups = [
  {
    name: "Workspace",
    items: [
      { name: "Overview", href: "/dashboard", icon: LayoutGrid },
      { name: "Ask Xem", href: "/", icon: Sparkles },
      { name: "Inbox", href: "/inbox", icon: Mail },
      { name: "Outbox", href: "/developer/logs/emails", icon: Send },
    ],
  },
  {
    name: "Create",
    items: [
      { name: "Campaigns", href: "/campaigns", icon: Mail },
      { name: "Newsletters", href: "/newsletters", icon: Newspaper },
      { name: "Automations", href: "/automations", icon: Workflow },
      { name: "Templates", href: "/templates", icon: PanelsTopLeft },
      { name: "Forms", href: "/forms", icon: FilePenLine },
    ],
  },
  {
    name: "Audience",
    items: [
      { name: "Contact lists", href: "/audience/lists", icon: Users },
      { name: "CRM", href: "/crm", icon: Users },
      { name: "Analytics", href: "/analytics", icon: ChartNoAxesCombined },
    ],
  },
];
const navigation = navigationGroups.flatMap((group) =>
  group.items.map((item) => ({ ...item, group: group.name })),
);
const gettingStarted = {
  name: "Getting started",
  href: "/onboarding",
  icon: CircleHelp,
  group: "Workspace",
};
const settingsNavigation = [
  { name: "Managed sending", href: "/settings/sending", icon: Send },
  { name: "SMTP senders", href: "/settings/smtp", icon: Server },
  { name: "IMAP mailboxes", href: "/settings/imap", icon: Mailbox },
  { name: "API keys", href: "/settings/api-keys", icon: KeyRound },
  { name: "Webhooks", href: "/settings/webhooks", icon: Webhook },
  { name: "Tags", href: "/audience/tags", icon: Tags },
  { name: "Team", href: "/team", icon: Users },
  { name: "Billing", href: "/settings/billing", icon: CreditCard },
  { name: "Account", href: "/settings", icon: UserRound },
];
const searchNavigation = [
  ...navigation,
  gettingStarted,
  ...settingsNavigation.map((item) => ({ ...item, group: "Settings" })),
];
const sharedHeadingPages = [
  "/campaigns",
  "/settings",
  "/settings/billing",
  "/settings/api-keys",
  "/settings/smtp",
  "/settings/imap",
  "/team",
];

export function AppShell({
  children,
  previewPage,
  onPreviewNavigate,
}: {
  children: ReactNode;
  previewPage?: string;
  onPreviewNavigate?: (page: string) => void;
}) {
  const pathname = usePathname();
  const router = useRouter();
  const { data: session } = useSession();
  const { theme, setTheme } = useTheme();
  const [mobile, setMobile] = useState(false);
  const [searchOpen, setSearchOpen] = useState(false);
  const [search, setSearch] = useState("");
  const [collapsed, setCollapsed] = useState(false);
  const searchResultsRef = useRef<HTMLDivElement>(null);
  const active = previewPage || pathname;

  useEffect(() => {
    function keydown(event: KeyboardEvent) {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === "k") {
        event.preventDefault();
        setSearchOpen((value) => !value);
      }
    }
    window.addEventListener("keydown", keydown);
    return () => window.removeEventListener("keydown", keydown);
  }, []);

  useEffect(() => {
    setMobile(false);
  }, [active]);

  useEffect(() => {
    if (!searchOpen) setSearch("");
  }, [searchOpen]);

  const current = navigation.find((item) =>
    item.href === "/"
      ? active === "/"
      : item.href === "/analytics"
        ? active.startsWith("/analytics") || active === "/audience/dashboard"
        : active === item.href || active.startsWith(`${item.href}/`),
  );
  const currentSetting = settingsNavigation.find(
    (item) =>
      active === item.href ||
      (item.href !== "/settings" && active.startsWith(`${item.href}/`)),
  );
  const currentPage =
    currentSetting?.name ||
    current?.name ||
    (active === "/onboarding" ? "Getting started" : "Workspace");
  const currentGroup = currentSetting
    ? "Settings"
    : current?.group || "Workspace";
  const results = searchNavigation.filter((item) =>
    `${item.name} ${item.group}`
      .toLowerCase()
      .includes(search.trim().toLowerCase()),
  );
  const name =
    session?.user?.name || (previewPage ? "Alex Morgan" : "Your account");
  const initials = name
    .split(" ")
    .filter(Boolean)
    .map((part) => part[0])
    .slice(0, 2)
    .join("");

  if (
    pathname.startsWith("/f/") ||
    pathname.startsWith("/auth/") ||
    (pathname === "/preview" && !previewPage)
  ) {
    return <>{children}</>;
  }

  const go = (href: string) => {
    setMobile(false);
    setSearchOpen(false);
    onPreviewNavigate ? onPreviewNavigate(href) : router.push(href);
  };
  const handleNavigation = (
    event: React.MouseEvent<HTMLAnchorElement>,
    href: string,
  ) => {
    if (onPreviewNavigate) {
      event.preventDefault();
      go(href);
    } else {
      setMobile(false);
    }
  };

  const sidebar = (isMobile = false) => {
    const compact = collapsed && !isMobile;
    return (
      <>
        <div className={styles.brandRow}>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <button
                type="button"
                className={styles.workspaceIdentity}
                aria-label="Workspace menu"
                title={compact ? "Xem workspace" : undefined}
              >
                <img
                  src="/android-chrome-512x512.png"
                  alt=""
                  className={styles.brandMark}
                />
                {!compact && (
                  <>
                    <span className={styles.workspaceIdentityText}>
                      <strong>Xem workspace</strong>
                      <small>{name}</small>
                    </span>
                    <ChevronsUpDown size={13} strokeWidth={1.6} />
                  </>
                )}
              </button>
            </DropdownMenuTrigger>
            <DropdownMenuContent
              side="bottom"
              align="start"
              className={styles.workspaceMenu}
            >
              <DropdownMenuItem onClick={() => go("/settings")}>
                <Settings size={15} />
                Workspace settings
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => go("/team")}>
                <Users size={15} />
                Team
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => go("/settings/billing")}>
                <LayoutGrid size={15} />
                Billing
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
          {!isMobile && !compact && (
            <button
              type="button"
              className={styles.iconButton}
              aria-label="Collapse sidebar"
              onClick={() => setCollapsed(true)}
            >
              <PanelLeftClose size={16} strokeWidth={1.6} />
            </button>
          )}
        </div>
        {compact && (
          <button
            type="button"
            className={cn(styles.iconButton, styles.expandButton)}
            aria-label="Expand sidebar"
            onClick={() => setCollapsed(false)}
          >
            <PanelLeftOpen size={17} strokeWidth={1.6} />
          </button>
        )}
        <button
          type="button"
          className={cn(styles.sidebarSearch, compact && styles.compactSearch)}
          aria-label="Search workspace"
          title={compact ? "Search workspace" : undefined}
          onClick={() => {
            setMobile(false);
            setSearchOpen(true);
          }}
        >
          <Search size={16} strokeWidth={1.65} />
          {!compact && (
            <>
              <span>Search</span>
              <kbd>⌘ K</kbd>
            </>
          )}
        </button>
        <nav
          className={styles.navigation}
          aria-label={isMobile ? "Mobile navigation" : "Main navigation"}
        >
          {navigationGroups.map((group) => (
            <div className={styles.navGroup} key={group.name}>
              {!compact && <p className={styles.groupLabel}>{group.name}</p>}
              {group.items.map(({ name: label, href, icon: Icon }) => (
                <Link
                  key={href}
                  href={href}
                  title={compact ? label : undefined}
                  aria-label={compact ? label : undefined}
                  aria-current={current?.href === href ? "page" : undefined}
                  onClick={(event) => handleNavigation(event, href)}
                  className={cn(
                    styles.navItem,
                    current?.href === href && styles.active,
                  )}
                >
                  <Icon size={15} strokeWidth={1.65} />
                  {!compact && <span>{label}</span>}
                </Link>
              ))}
            </div>
          ))}
        </nav>
        <div className={styles.sidebarBottom}>
          <Link
            href="/onboarding"
            title={compact ? "Getting started" : undefined}
            aria-label={compact ? "Getting started" : undefined}
            aria-current={active === "/onboarding" ? "page" : undefined}
            onClick={(event) => handleNavigation(event, "/onboarding")}
            className={cn(
              styles.navItem,
              styles.onboarding,
              active === "/onboarding" && styles.active,
            )}
          >
            <CircleHelp size={17} strokeWidth={1.65} />
            {!compact && (
              <>
                <span>Getting started</span>
                <ArrowUpRight size={13} className={styles.trailingIcon} />
              </>
            )}
          </Link>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <button
                type="button"
                className={cn(
                  styles.navItem,
                  styles.settingsButton,
                  currentSetting && styles.active,
                )}
                title={compact ? "Settings" : undefined}
                aria-label="Settings and workspace"
              >
                <Settings size={17} strokeWidth={1.65} />
                {!compact && (
                  <>
                    <span>Settings</span>
                    <ChevronDown size={13} className={styles.trailingIcon} />
                  </>
                )}
              </button>
            </DropdownMenuTrigger>
            <DropdownMenuContent
              side={isMobile ? "top" : "right"}
              align="end"
              sideOffset={8}
              collisionPadding={16}
              className={styles.workspaceMenu}
            >
              {settingsNavigation.map(({ name: label, href, icon: Icon }) => (
                <DropdownMenuItem
                  key={href}
                  onClick={() => go(href)}
                  className={
                    currentSetting?.href === href
                      ? styles.menuActive
                      : undefined
                  }
                >
                  <Icon size={15} strokeWidth={1.65} />
                  {label}
                </DropdownMenuItem>
              ))}
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </>
    );
  };

  return (
    <div
      className={cn(
        "product-frame",
        styles.frame,
        collapsed && styles.collapsed,
      )}
    >
      <aside className={cn("product-sidebar", styles.sidebar)}>
        {sidebar()}
      </aside>
      <Sheet open={mobile} onOpenChange={setMobile}>
        <SheetContent side="left" className={styles.mobileSidebar}>
          <SheetTitle className="sr-only">Workspace navigation</SheetTitle>
          <SheetDescription className="sr-only">
            Open a page or change your workspace settings.
          </SheetDescription>
          {sidebar(true)}
        </SheetContent>
      </Sheet>
      <div className={cn("product-main", styles.main)}>
        <header className={cn("product-topbar", styles.topbar)}>
          <button
            type="button"
            className={cn(styles.iconButton, styles.mobileMenu)}
            aria-label="Open navigation"
            onClick={() => setMobile(true)}
          >
            <Menu size={19} strokeWidth={1.65} />
          </button>
          <nav className={styles.breadcrumb} aria-label="Breadcrumb">
            <span className={styles.breadcrumbGroup}>{currentGroup}</span>
            <span className={styles.breadcrumbSeparator} aria-hidden="true">
              /
            </span>
            <span aria-current="page">{currentPage}</span>
          </nav>
          <div className={styles.topbarActions}>
            <button
              type="button"
              aria-label="Search your workspace"
              className={styles.searchTrigger}
              onClick={() => setSearchOpen(true)}
            >
              <Search size={16} strokeWidth={1.65} />
              <span>Search</span>
              <kbd>⌘ K</kbd>
            </button>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <button
                  type="button"
                  aria-label={`Account menu for ${name}`}
                  className={styles.accountTrigger}
                >
                  <span className={styles.avatar}>{initials}</span>
                  <ChevronDown size={13} strokeWidth={1.65} />
                </button>
              </DropdownMenuTrigger>
              <DropdownMenuContent
                side="bottom"
                align="end"
                sideOffset={8}
                collisionPadding={12}
                className={styles.workspaceMenu}
              >
                <div className={styles.accountInfo}>
                  <p>{name}</p>
                  <span>{session?.user?.email || "Your Xem workspace"}</span>
                </div>
                <DropdownMenuItem onClick={() => go("/settings")}>
                  <Settings size={15} strokeWidth={1.65} />
                  Account settings
                </DropdownMenuItem>
                <DropdownMenuSub>
                  <DropdownMenuSubTrigger className={styles.appearanceTrigger}>
                    <Monitor size={15} strokeWidth={1.65} />
                    Appearance
                  </DropdownMenuSubTrigger>
                  <DropdownMenuSubContent className={styles.workspaceMenu}>
                    <DropdownMenuRadioGroup
                      value={theme}
                      onValueChange={setTheme}
                    >
                      <DropdownMenuRadioItem
                        value="dark"
                        className={styles.appearanceChoice}
                      >
                        <Moon size={15} strokeWidth={1.65} />
                        Dark
                      </DropdownMenuRadioItem>
                      <DropdownMenuRadioItem
                        value="light"
                        className={styles.appearanceChoice}
                      >
                        <Sun size={15} strokeWidth={1.65} />
                        Light
                      </DropdownMenuRadioItem>
                      <DropdownMenuRadioItem
                        value="system"
                        className={styles.appearanceChoice}
                      >
                        <Monitor size={15} strokeWidth={1.65} />
                        System
                      </DropdownMenuRadioItem>
                    </DropdownMenuRadioGroup>
                  </DropdownMenuSubContent>
                </DropdownMenuSub>
                <DropdownMenuSeparator />
                <DropdownMenuItem
                  onClick={() => signOut({ callbackUrl: "/auth/login" })}
                >
                  <LogOut size={15} strokeWidth={1.65} />
                  Sign out
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </header>
        <main
          className={cn(
            "product-content",
            styles.content,
            ["/", "/inbox", "/developer/logs/emails"].includes(active) &&
              styles.assistantContent,
          )}
          data-page={active}
        >
          {sharedHeadingPages.includes(active) && !previewPage && <AppHeader />}
          {children}
        </main>
      </div>
      <Modal
        open={searchOpen}
        onOpenChange={setSearchOpen}
        title="Search workspace"
        description="Open a page, tool, or setting."
      >
        <div className={styles.searchField}>
          <Search size={17} strokeWidth={1.65} />
          <input
            autoFocus
            aria-label="Search workspace pages"
            placeholder="Where would you like to go?"
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            onKeyDown={(event) => {
              if (event.key === "ArrowDown") {
                event.preventDefault();
                searchResultsRef.current?.querySelector("button")?.focus();
              }
            }}
          />
        </div>
        <div
          ref={searchResultsRef}
          className={styles.searchResults}
          onKeyDown={(event) => {
            if (!["ArrowDown", "ArrowUp", "Home", "End"].includes(event.key))
              return;
            const buttons = Array.from(
              event.currentTarget.querySelectorAll("button"),
            );
            const index = buttons.indexOf(
              document.activeElement as HTMLButtonElement,
            );
            const nextIndex =
              event.key === "Home"
                ? 0
                : event.key === "End"
                  ? buttons.length - 1
                  : (index +
                      (event.key === "ArrowDown" ? 1 : -1) +
                      buttons.length) %
                    buttons.length;
            event.preventDefault();
            buttons[nextIndex]?.focus();
          }}
        >
          {results.map(({ name: label, href, icon: Icon, group }) => (
            <button key={href} type="button" onClick={() => go(href)}>
              <Icon size={15} strokeWidth={1.65} />
              <span>{label}</span>
              <small>{group}</small>
              <ArrowUpRight size={14} strokeWidth={1.65} />
            </button>
          ))}
          {results.length === 0 && (
            <p className={styles.searchEmpty}>
              No pages match “{search}”. Try another name.
            </p>
          )}
        </div>
      </Modal>
    </div>
  );
}
