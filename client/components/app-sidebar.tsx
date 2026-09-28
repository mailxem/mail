"use client";

import { cn } from "@/lib/utils";
import { ChevronDown, ChevronUp, Settings } from "lucide-react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import React, { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Sidebar } from "./ui/sidebar";
import { Sheet, SheetContent, SheetHeader, SheetTrigger } from "./ui/sheet";
import { ComposeEmail } from "./compose-email";
import { useSession } from "next-auth/react";
import Intercom from "@intercom/messenger-js-sdk";
import { toast } from "sonner";
import { useApi } from "@/hooks/use-api";

type SubMenuItem = {
  name: string;
  href: string;
};

type NavItem = {
  name: string;
  href: string;
  icon: React.ReactNode;
  subItems?: SubMenuItem[];
};

const navItems: NavItem[] = [
  {
    name: "Home",
    href: "/",
    icon: (
      <img
        src="https://img.icons8.com/quill/100/home.png"
        className="w-6 dark:invert invert-0"
      />
    ),
  },
  {
    name: "Inbox",
    href: "/inbox",
    icon: (
      <img
        src="https://img.icons8.com/quill/100/inbox.png"
        className="w-6 dark:invert invert-0"
      />
    ),
  },
  {
    name: "Outbox",
    href: "/developer/logs/emails",
    icon: (
      <img
        src="https://img.icons8.com/quill/100/filled-sent.png"
        className="w-6 dark:invert invert-0"
      />
    ),
  },
  {
    name: "Campaigns",
    href: "/campaigns",
    icon: (
      <img
        alt="svgImg"
        className="w-6 dark:invert invert-0"
        src="data:image/svg+xml;base64,PD94bWwgdmVyc2lvbj0iMS4wIj8+PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciICB2aWV3Qm94PSIwIDAgMTI4IDEyOCIgd2lkdGg9IjEyOHB4IiBoZWlnaHQ9IjEyOHB4Ij4gICAgPHBhdGggZD0iTSA4OCAxNCBDIDg0LjIgMTQgODEuMDk5NjA5IDE3LjEwMDM5MSA4MS4wOTk2MDkgMjAuOTAwMzkxIEwgODEuMDk5NjA5IDI0LjE5OTIxOSBDIDc2LjY5OTYwOSAzNC42OTkyMTkgNjYuNjAwNzgxIDQxLjQwMDM5MSA1NS4zMDA3ODEgNDEuNDAwMzkxIEwgNTUuMTk5MjE5IDQxLjQwMDM5MSBDIDU0LjA5OTIxOSAzOS4xMDAzOTEgNTEuNyAzNy41IDQ5IDM3LjUgTCAyNS41OTk2MDkgMzcuNSBDIDIyLjg5OTYwOSAzNy41IDIwLjUwMDM5MSAzOS4xMDAzOTEgMTkuNDAwMzkxIDQxLjQwMDM5MSBMIDEzLjkwMDM5MSA0MS40MDAzOTEgQyA3LjkwMDM5MDYgNDEuNDAwMzkxIDMgNDYuMzAwNzgxIDMgNTIuMzAwNzgxIEwgMyA2OCBDIDMgNzQgNy45MDA3ODEyIDc4LjgwMDc4MSAxMy44MDA3ODEgNzguODAwNzgxIEwgMTkuMzAwNzgxIDc4LjgwMDc4MSBMIDE5LjMwMDc4MSA3OC42OTkyMTkgQyAyMC40MDA3ODEgODAuOTk5MjE5IDIyLjggODIuNTk5NjA5IDI1LjUgODIuNTk5NjA5IEwgMjYuNDAwMzkxIDgyLjU5OTYwOSBMIDI2LjQwMDM5MSAxMDkuNDAwMzkgQyAyNi40MDAzOTEgMTExLjkwMDM5IDI4LjQgMTE0IDMxIDExNCBMIDM0LjMwMDc4MSAxMTQgQyAzOS43MDA3ODEgMTE0IDQ0LjQgMTA5LjkgNDUgMTA0LjUgTCA0Ny42OTkyMTkgODIuNjk5MjE5IEwgNDkgODIuNjk5MjE5IEMgNTEuNyA4Mi42OTkyMTkgNTQuMDk5MjE5IDgxLjEwMDc4MSA1NS4xOTkyMTkgNzguODAwNzgxIEwgNTUuMzAwNzgxIDc4LjgwMDc4MSBDIDY2LjcwMDc4MSA3OC44MDA3ODEgNzYuNjk5NjA5IDg1LjUgODEuMDk5NjA5IDk2IEwgODEuMDk5NjA5IDk5LjMwMDc4MSBDIDgxLjA5OTYwOSAxMDMuMTAwNzggODQuMiAxMDYuMTk5MjIgODggMTA2LjE5OTIyIEMgOTEuOCAxMDYuMTk5MjIgOTQuOTAwMzkxIDEwMy4xMDA3OCA5NC45MDAzOTEgOTkuMzAwNzgxIEwgOTQuOTAwMzkxIDIwLjkwMDM5MSBDIDk0LjkwMDM5MSAxNy4xMDAzOTEgOTEuOCAxNCA4OCAxNCB6IE0gODguMDk5NjA5IDIwIEMgODguNTk5NjA5IDIwIDg5IDIwLjQwMDM5MSA4OSAyMC45MDAzOTEgTCA4OSA5MS4zMDA3ODEgQyA4OSA5MS44MDA3ODEgODguNTk5NjA5IDkyLjE5OTIxOSA4OC4wOTk2MDkgOTIuMTk5MjE5IEMgODcuNTk5NjA5IDkyLjE5OTIxOSA4Ny4xOTkyMTkgOTEuODAwNzgxIDg3LjE5OTIxOSA5MS4zMDA3ODEgTCA4Ny4xOTkyMTkgODcuNDAwMzkxIEwgODcuMTk5MjE5IDI0LjgwMDc4MSBMIDg3LjE5OTIxOSAyMC45MDAzOTEgQyA4Ny4xOTkyMTkgMjAuNDAwMzkxIDg3LjU5OTYwOSAyMCA4OC4wOTk2MDkgMjAgeiBNIDExNC43NTk3NyAzMC41NTQ2ODggQyAxMTQuMzY5MTQgMzAuNTg3NSAxMTMuOTc0NjEgMzAuNzAwMzkxIDExMy41OTk2MSAzMC45MDAzOTEgTCAxMDMuNTk5NjEgMzUuOTAwMzkxIEMgMTAyLjQ5OTYxIDM2LjQwMDM5MSAxMDEuOTAwMzkgMzcuNDk5NjA5IDEwMS45MDAzOSAzOC41OTk2MDkgQyAxMDEuOTAwMzkgMzkuMDk5NjA5IDEwMS45OTkyMiAzOS41MDAzOTEgMTAyLjE5OTIyIDM5LjkwMDM5MSBDIDEwMy4wOTkyMiA0MS4zMDAzOTEgMTA0LjkwMDc4IDQxLjg5OTIxOSAxMDYuMzAwNzggNDEuMTk5MjE5IEwgMTE2LjMwMDc4IDM2LjE5OTIxOSBDIDExNy44MDA3OCAzNS40OTkyMTkgMTE4LjM5OTYxIDMzLjY5OTIxOSAxMTcuNTk5NjEgMzIuMTk5MjE5IEMgMTE3LjA3NDYxIDMxLjA3NDIxOSAxMTUuOTMxNjQgMzAuNDU2MjUgMTE0Ljc1OTc3IDMwLjU1NDY4OCB6IE0gODEuMDk5NjA5IDM1LjUgTCA4MS4wOTk2MDkgNzcuNTk5NjA5IEMgNzQuODk5NjA5IDcwLjI5OTYwOSA2NS44MDAzOTEgNjUuODk5MjE5IDU1LjkwMDM5MSA2NS42OTkyMTkgTCA1NS45MDAzOTEgNDcuNDAwMzkxIEMgNjUuODAwMzkxIDQ3LjIwMDM5MSA3NC44OTk2MDkgNDIuOCA4MS4wOTk2MDkgMzUuNSB6IE0gMjUuNTk5NjA5IDQzLjUgTCA0OSA0My41IEMgNDkuNSA0My41IDQ5LjkwMDM5MSA0My45MDAzOTEgNDkuOTAwMzkxIDQ0LjQwMDM5MSBMIDQ5LjkwMDM5MSA3NS42OTkyMTkgQyA0OS45MDAzOTEgNzYuMTk5MjE5IDQ5LjUgNzYuNTk5NjA5IDQ5IDc2LjU5OTYwOSBMIDQ1LjA5OTYwOSA3Ni41OTk2MDkgTCAyOS41IDc2LjU5OTYwOSBMIDI1LjU5OTYwOSA3Ni41OTk2MDkgQyAyNS4wOTk2MDkgNzYuNTk5NjA5IDI0LjY5OTIxOSA3Ni4xOTkyMTkgMjQuNjk5MjE5IDc1LjY5OTIxOSBMIDI0LjY5OTIxOSA0NC40MDAzOTEgQyAyNC42OTkyMTkgNDMuOTAwMzkxIDI1LjA5OTYwOSA0My41IDI1LjU5OTYwOSA0My41IHogTSAxMy44MDA3ODEgNDcuNSBMIDE4LjU5OTYwOSA0Ny41IEwgMTguNTk5NjA5IDcyLjgwMDc4MSBMIDEzLjgwMDc4MSA3Mi44MDA3ODEgQyAxMS4yMDA3ODEgNzIuNzAwNzgxIDkgNzAuNjAwMzkxIDkgNjcuOTAwMzkxIEwgOSA1Mi4zMDA3ODEgQyA5IDQ5LjYwMDc4MSAxMS4yMDA3ODEgNDcuNSAxMy44MDA3ODEgNDcuNSB6IE0gMTA1IDU2LjUgQyAxMDMuMyA1Ni41IDEwMiA1Ny44IDEwMiA1OS41IEMgMTAyIDYxLjIgMTAzLjMgNjIuNSAxMDUgNjIuNSBMIDEyMCA2Mi41IEMgMTIxLjcgNjIuNSAxMjMgNjEuMiAxMjMgNTkuNSBDIDEyMyA1Ny44IDEyMS43IDU2LjUgMTIwIDU2LjUgTCAxMDUgNTYuNSB6IE0gMTA1LjE0MDYyIDc3LjUxMTcxOSBDIDEwMy45Njg3NSA3Ny40NTA3ODEgMTAyLjgyNTc4IDc4LjA0OTYwOSAxMDIuMzAwNzggNzkuMDk5NjA5IEMgMTAyLjEwMDc4IDc5LjQ5OTYwOSAxMDIgODAuMDAwMzkxIDEwMiA4MC40MDAzOTEgQyAxMDIgODEuNTAwMzkxIDEwMi41OTkyMiA4Mi41OTk2MDkgMTAzLjY5OTIyIDgzLjA5OTYwOSBMIDExMy42OTkyMiA4OC4wOTk2MDkgQyAxMTUuMTk5MjIgODguNzk5NjA5IDExNi45OTkyMiA4OC4yMDA3ODEgMTE3LjY5OTIyIDg2LjgwMDc4MSBDIDExOC4zOTkyMiA4NS40MDA3ODEgMTE3LjgwMDc4IDgzLjYwMDc4MSAxMTYuMzAwNzggODIuODAwNzgxIEwgMTA2LjMwMDc4IDc3LjgwMDc4MSBDIDEwNS45MjU3OCA3Ny42MjU3ODEgMTA1LjUzMTI1IDc3LjUzMjAzMSAxMDUuMTQwNjIgNzcuNTExNzE5IHogTSAzMi41IDg4LjY5OTIxOSBMIDQwLjY5OTIxOSA4OC42OTkyMTkgTCAzOS4wOTk2MDkgMTAzLjgwMDc4IEMgMzguNzk5NjA5IDEwNi4yMDA3OCAzNi43MDA3ODEgMTA4IDM0LjMwMDc4MSAxMDggTCAzMi41IDEwOCBMIDMyLjUgODguNjk5MjE5IHoiLz48L3N2Zz4="
      />
    ),
    subItems: [
      { name: "All campaigns", href: "/campaigns" },
      { name: "Templates", href: "/templates" },
    ],
  },
  {
    name: "Automations",
    href: "/automations",
    icon: (
      <img
        alt="svgImg"
        className="w-6 dark:invert invert-0"
        src="data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciICB2aWV3Qm94PSIwIDAgMzAgMzAiIHdpZHRoPSIzMHB4IiBoZWlnaHQ9IjMwcHgiPjxwYXRoIGQ9Ik0yMC45OTcgMTkuOTk0QTQuMDAzIDQuMDAzIDAgMTAyMC45OTcgMjggNC4wMDMgNC4wMDMgMCAxMDIwLjk5NyAxOS45OTR6TTI0LjExNCAzLjI0OGMtLjM3NS0uMzE0LS45MjQtLjMxMy0xLjI5Ni4wMDhsLTMuNDU4IDIuOTc5Yy0uNDE4LjM2LS40NjUuOTkyLS4xMDUgMS40MXMuOTkxLjQ2NiAxLjQxMS4xMDVsMS44NDctMS41OTF2NS41ODNjMCAyLjE4OS0xLjk0OSAyLjMyNy0yLjMzMyAyLjMzM0g5LjM0NWMtMy4xNzkgMC00LjMzMyAyLjU5Mi00LjMzMyA0LjMzNHY3LjE2NmMwIC41NTMuNDQ4IDEgMSAxczEtLjQ0NyAxLTF2LTcuMTU5Yy4wMDUtLjM5MS4xNDQtMi4zNDEgMi4zMzMtMi4zNDFoMTAuODMzYzEuNzQyIDAgNC4zMzMtMS4xNTQgNC4zMzMtNC4zMzNWNi4xOTdsMS44NTYgMS41NjJjLjE4OC4xNTcuNDE2LjIzNC42NDMuMjM0LjI4NSAwIC41NjgtLjEyMS43NjYtLjM1Ni4zNTUtLjQyMy4zMDEtMS4wNTMtLjEyMi0xLjQwOUwyNC4xMTQgMy4yNDh6TTEwLjcwMSAzLjI2NmMtLjQtLjM5LTEuMDMtLjM5LTEuNDIgMGwtMS4yOSAxLjMtMS4yOS0xLjNjLS40LS4zOS0xLjAzLS4zOS0xLjQyIDAtLjM5LjM5LS4zOSAxLjAzIDAgMS40MmwxLjI5IDEuMjktMS4yOSAxLjI5Yy0uMzkuMzktLjM5IDEuMDMgMCAxLjQyLjIuMTkuNDUuMjkuNzEuMjkuMjUgMCAuNTEtLjEuNzEtLjI5bDEuMjktMS4zIDEuMjkgMS4zYy4yLjE5LjQ1LjI5LjcxLjI5LjI1IDAgLjUxLS4xLjcxLS4yOS4zOS0uMzkuMzktMS4wMyAwLTEuNDJsLTEuMy0xLjI5IDEuMy0xLjI5QzExLjA5MSA0LjI5NiAxMS4wOTEgMy42NTYgMTAuNzAxIDMuMjY2eiIvPjwvc3ZnPg=="
      />
    ),
    subItems: [
      { name: "Overview", href: "/automations" },
      { name: "All journeys", href: "/automations/journeys" },
      { name: "Pre-built journeys", href: "/automations/pre-built" },
      { name: "Transactional email", href: "/automations/transactional" },
    ],
  },
  {
    name: "Audience",
    href: "/audience",
    icon: (
      <img
        alt="svgImg"
        className="w-6 dark:invert invert-0"
        src="data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciICB2aWV3Qm94PSIwIDAgNTAgNTAiIHdpZHRoPSI1MHB4IiBoZWlnaHQ9IjUwcHgiPjxwYXRoIGQ9Ik0gMjEuNjk5MjE5IDMwIEMgMjEuODk4NDM4IDI5LjY5OTIxOSAyMiAyOS4zOTg0MzggMjIgMjkgQyAyMiAyNy44OTg0MzggMjEuMTAxNTYzIDI3IDIwIDI3IEMgMTguODk4NDM4IDI3IDE4IDI3Ljg5ODQzOCAxOCAyOSBDIDE4IDI5LjM5ODQzOCAxOC4xMDE1NjMgMjkuNjk5MjE5IDE4LjMwMDc4MSAzMCBaIE0gMzEuNjk5MjE5IDMwIEMgMzEuODk4NDM4IDI5LjY5OTIxOSAzMiAyOS4zOTg0MzggMzIgMjkgQyAzMiAyNy44OTg0MzggMzEuMTAxNTYzIDI3IDMwIDI3IEMgMjguODk4NDM4IDI3IDI4IDI3Ljg5ODQzOCAyOCAyOSBDIDI4IDI5LjM5ODQzOCAyOC4xMDE1NjMgMjkuNjk5MjE5IDI4LjMwMDc4MSAzMCBaIE0gMzEgMzQgTCAzMSAzNSBDIDMxIDM4LjM2MzI4MSAyOC4zNjMyODEgNDEgMjUgNDEgQyAyMS42MzY3MTkgNDEgMTkgMzguMzYzMjgxIDE5IDM1IEwgMTkgMzQgWiBNIDE3IDI2LjMwMDc4MSBDIDE2Ljc2NTYyNSAyNi4zMDA3ODEgMTYuNTMxMjUgMjYuMjE4NzUgMTYuMzQzNzUgMjYuMDUwNzgxIEMgMTUuOTI1NzgxIDI1LjY4NzUgMTUuODgyODEzIDI1LjA1ODU5NCAxNi4yNDYwOTQgMjQuNjQwNjI1IEMgMTcuMjAzMTI1IDIzLjU1MDc4MSAxOC42MDU0NjkgMjIuODk4NDM4IDIwIDIyLjg5ODQzOCBDIDIwLjgwODU5NCAyMi44OTg0MzggMjEuNjUyMzQ0IDIzLjEwOTM3NSAyMi40NDUzMTMgMjMuNTAzOTA2IEMgMjIuOTQxNDA2IDIzLjc1MzkwNiAyMy4xNDA2MjUgMjQuMzUxNTYzIDIyLjg5NDUzMSAyNC44NDc2NTYgQyAyMi42NDg0MzggMjUuMzM5ODQ0IDIyLjA1MDc4MSAyNS41MzkwNjMgMjEuNTU0Njg4IDI1LjI5Mjk2OSBDIDIxLjA0Mjk2OSAyNS4wMzkwNjMgMjAuNDkyMTg4IDI0Ljg5ODQzOCAyMCAyNC44OTg0MzggQyAxOS4xODM1OTQgMjQuODk4NDM4IDE4LjMyNDIxOSAyNS4zMDQ2ODggMTcuNzUzOTA2IDI1Ljk2MDkzOCBDIDE3LjU1NDY4OCAyNi4xODM1OTQgMTcuMjc3MzQ0IDI2LjMwMDc4MSAxNyAyNi4zMDA3ODEgWiBNIDMzIDI2LjM5ODQzOCBDIDMyLjcwMzEyNSAyNi4zOTg0MzggMzIuNDA2MjUgMjYuMjY1NjI1IDMyLjIxMDkzOCAyNi4wMTU2MjUgQyAzMS40Mzc1IDI1LjAxOTUzMSAyOS44OTA2MjUgMjQuNjcxODc1IDI4LjQ0NTMxMyAyNS4zOTQ1MzEgQyAyNy45NTMxMjUgMjUuNjQwNjI1IDI3LjM1MTU2MyAyNS40NDE0MDYgMjcuMTA1NDY5IDI0Ljk0NTMxMyBDIDI2Ljg1OTM3NSAyNC40NTMxMjUgMjcuMDU4NTk0IDIzLjg1MTU2MyAyNy41NTA3ODEgMjMuNjA1NDY5IEMgMjkuODM5ODQ0IDIyLjQ2NDg0NCAzMi40Mzc1IDIzLjA1MDc4MSAzMy43ODkwNjMgMjQuNzg1MTU2IEMgMzQuMTI4OTA2IDI1LjIyMjY1NiAzNC4wNTA3ODEgMjUuODUxNTYzIDMzLjYxMzI4MSAyNi4xODc1IEMgMzMuNDMzNTk0IDI2LjMzMjAzMSAzMy4yMTQ4NDQgMjYuMzk4NDM4IDMzIDI2LjM5ODQzOCBaIE0gMTEgNyBDIDEwLjY3OTY4OCA3IDEwLjM1OTM3NSA2Ljg0Mzc1IDEwLjE2Nzk2OSA2LjU1NDY4OCBMIDguMTY3OTY5IDMuNTU0Njg4IEMgNy44NjMyODEgMy4wOTM3NSA3Ljk4NDM3NSAyLjQ3MjY1NiA4LjQ0NTMxMyAyLjE2Nzk2OSBDIDguOTAyMzQ0IDEuODU5Mzc1IDkuNTIzNDM4IDEuOTg0Mzc1IDkuODMyMDMxIDIuNDQ1MzEzIEwgMTEuODMyMDMxIDUuNDQ1MzEzIEMgMTIuMTM2NzE5IDUuOTAyMzQ0IDEyLjAxNTYyNSA2LjUyNzM0NCAxMS41NTQ2ODggNi44MzIwMzEgQyAxMS4zODI4MTMgNi45NDUzMTMgMTEuMTkxNDA2IDcgMTEgNyBaIE0gOCAxMiBDIDcuODc1IDEyIDcuNzUgMTEuOTc2NTYzIDcuNjI4OTA2IDExLjkyOTY4OCBMIDIuNjI4OTA2IDkuOTI5Njg4IEMgMi4xMTcxODggOS43MjI2NTYgMS44NjcxODggOS4xNDA2MjUgMi4wNzAzMTMgOC42Mjg5MDYgQyAyLjI3NzM0NCA4LjExMzI4MSAyLjg1NTQ2OSA3Ljg2NzE4OCAzLjM3MTA5NCA4LjA3MDMxMyBMIDguMzcxMDk0IDEwLjA3MDMxMyBDIDguODg2NzE5IDEwLjI3NzM0NCA5LjEzMjgxMyAxMC44NTkzNzUgOC45Mjk2ODggMTEuMzcxMDk0IEMgOC43NzM0MzggMTEuNzYxNzE5IDguMzk0NTMxIDEyIDggMTIgWiBNIDcgMTggTCAzIDE4IEMgMi40NDkyMTkgMTggMiAxNy41NTA3ODEgMiAxNyBDIDIgMTYuNDQ5MjE5IDIuNDQ5MjE5IDE2IDMgMTYgTCA3IDE2IEMgNy41NTA3ODEgMTYgOCAxNi40NDkyMTkgOCAxNyBDIDggMTcuNTUwNzgxIDcuNTUwNzgxIDE4IDcgMTggWiBNIDM5IDcgQyAzOC44MDg1OTQgNyAzOC42MTcxODggNi45NDUzMTMgMzguNDQ1MzEzIDYuODMyMDMxIEMgMzcuOTg0Mzc1IDYuNTI3MzQ0IDM3Ljg1OTM3NSA1LjkwNjI1IDM4LjE2Nzk2OSA1LjQ0NTMxMyBMIDQwLjE2Nzk2OSAyLjQ0NTMxMyBDIDQwLjQ3MjY1NiAxLjk4NDM3NSA0MS4wOTM3NSAxLjg1OTM3NSA0MS41NTQ2ODggMi4xNjc5NjkgQyA0Mi4wMTU2MjUgMi40NzI2NTYgNDIuMTQwNjI1IDMuMDkzNzUgNDEuODMyMDMxIDMuNTU0Njg4IEwgMzkuODMyMDMxIDYuNTU0Njg4IEMgMzkuNjQwNjI1IDYuODQzNzUgMzkuMzIwMzEzIDcgMzkgNyBaIE0gNDIgMTIgQyA0MS42MDU0NjkgMTIgNDEuMjI2NTYzIDExLjc2MTcxOSA0MS4wNzAzMTMgMTEuMzcxMDk0IEMgNDAuODY3MTg4IDEwLjg1OTM3NSA0MS4xMTcxODggMTAuMjc3MzQ0IDQxLjYyODkwNiAxMC4wNzAzMTMgTCA0Ni42Mjg5MDYgOC4wNzAzMTMgQyA0Ny4xMzY3MTkgNy44NjcxODggNDcuNzIyNjU2IDguMTEzMjgxIDQ3LjkyOTY4OCA4LjYyODkwNiBDIDQ4LjEzMjgxMyA5LjE0MDYyNSA0Ny44ODI4MTMgOS43MjI2NTYgNDcuMzcxMDk0IDkuOTI5Njg4IEwgNDIuMzcxMDk0IDExLjkyOTY4OCBDIDQyLjI1IDExLjk3NjU2MyA0Mi4xMjUgMTIgNDIgMTIgWiBNIDQ3IDE4IEwgNDMgMTggQyA0Mi40NDUzMTMgMTggNDIgMTcuNTUwNzgxIDQyIDE3IEMgNDIgMTYuNDQ5MjE5IDQyLjQ0NTMxMyAxNiA0MyAxNiBMIDQ3IDE2IEMgNDcuNTU0Njg4IDE2IDQ4IDE2LjQ0OTIxOSA0OCAxNyBDIDQ4IDE3LjU1MDc4MSA0Ny41NTQ2ODggMTggNDcgMTggWiBNIDQxIDI1LjY0ODQzOCBMIDQxIDIwLjUgQyA0MSAxOS4wMTE3MTkgNDAuODk0NTMxIDE3LjY3MTg3NSA0MC42ODM1OTQgMTYuNTE5NTMxIEwgNDAuNjYwMTU2IDE2LjQyMTg3NSBDIDM4LjA3NDIxOSA3LjQ4MDQ2OSAzMS41NzQyMTkgNi40MTQwNjMgMjguNDYwOTM4IDYuMDQ2ODc1IEwgMjYuNDQxNDA2IDQuMDIzNDM4IEwgMjYuMDYyNSA0LjAwMzkwNiBDIDI1Ljc2MTcxOSAzLjk4MDQ2OSAxOC42MDE1NjMgMy41OTM3NSAxMy42MDE1NjMgOC4yODkwNjMgQyAxMC41NDY4NzUgMTEuMTU2MjUgOSAxNS4yNjU2MjUgOSAyMC41IEwgOSAyNS42NDg0MzggQyA4LjYxMzI4MSAyNi4xODc1IDggMjcuMzIwMzEzIDggMjkuMDk3NjU2IEMgOCAzMS45Njg3NSAxMC41MTE3MTkgMzMuMzc4OTA2IDExLjYwMTU2MyAzMy44NTkzNzUgQyAxMi40NTMxMjUgMzYuMjY1NjI1IDE2LjcyMjY1NiA0NyAyNSA0NyBDIDMzLjI3NzM0NCA0NyAzNy41NDY4NzUgMzYuMjY1NjI1IDM4LjM5ODQzOCAzMy44NTkzNzUgQyAzOS40ODgyODEgMzMuMzgyODEzIDQyIDMxLjk2ODc1IDQyIDI5LjEwMTU2MyBDIDQyIDI3LjMyMDMxMyA0MS4zODY3MTkgMjYuMTg3NSA0MSAyNS42NDg0MzggWiBNIDM3LjI5Mjk2OSAzMi4xNDg0MzggTCAzNi44MDQ2ODggMzIuMzA0Njg4IEwgMzYuNjQ4NDM4IDMyLjc5Njg3NSBDIDM2LjYwNTQ2OSAzMi45MTc5NjkgMzIuNjc5Njg4IDQ1IDI1IDQ1IEMgMTcuMzM1OTM4IDQ1IDEzLjM5MDYyNSAzMi45MTc5NjkgMTMuMzUxNTYzIDMyLjc5Njg3NSBMIDEzLjE5NTMxMyAzMi4zMDg1OTQgTCAxMi43MDcwMzEgMzIuMTQ4NDM4IEMgMTIuNjc5Njg4IDMyLjE0MDYyNSAxMCAzMS4yNDYwOTQgMTAgMjkuMTAxNTYzIEMgMTAgMjguMDM5MDYzIDEwLjI5Njg3NSAyNy4zNTU0NjkgMTAuNTAzOTA2IDI3IEwgMTIuMzk4NDM4IDI3IEMgMTIuNjk5MjE5IDI2Ljg5ODQzOCAxMyAyNi42MDE1NjMgMTMgMjYuMTk5MjE5IEwgMTQgMjAgQyAxNi4zMDA3ODEgMTkuNjk5MjE5IDI0IDE5IDI4IDE0IEMgMzAgMTggMzQuNjAxNTYzIDE5LjUgMzYgMjAgTCAzNyAyNi4xOTkyMTkgQyAzNy4xMDE1NjMgMjYuNjk5MjE5IDM3LjUgMjcgMzggMjcgTCAzOS40OTIxODggMjcgQyAzOS43MDcwMzEgMjcuMzU5Mzc1IDQwIDI4LjA0Mjk2OSA0MCAyOS4xMDE1NjMgQyA0MCAzMS4yMjY1NjMgMzcuMzE2NDA2IDMyLjE0MDYyNSAzNy4yOTI5NjkgMzIuMTQ4NDM4IFoiLz48L3N2Zz4="
      />
    ),
    subItems: [
      { name: "All contacts", href: "/audience/lists" },
      { name: "Audience dashboard", href: "/audience/dashboard" },
      { name: "Signup forms", href: "/forms" },
      { name: "Tags", href: "/audience/tags" },
    ],
  },
  {
    name: "Analytics",
    href: "/analytics",
    icon: (
      <img
        alt="svgImg"
        className="w-6 dark:invert invert-0"
        src="https://img.icons8.com/dotty/80/total-sales-1.png"
      />
    ),
    subItems: [
      { name: "Overview", href: "/analytics" },
      { name: "Campaigns", href: "/analytics/campaigns" },
      { name: "Team Performance", href: "/analytics/team" },
      { name: "Click Heatmaps", href: "/analytics/heatmap" },
      { name: "Audience Insights", href: "/analytics/audience" },
      { name: "Trend Analysis", href: "/analytics/trends" },
    ],
  },
  {
    name: "Developer",
    href: "/developer",
    icon: (
      <img
        src="https://img.icons8.com/ios-filled/500/einstein.png"
        className="w-6 dark:invert invert-0"
      />
    ),
    subItems: [
      { name: "API Keys", href: "/settings/api-keys" },
      { name: "Webhooks", href: "/developer/webhooks" },
      { name: "Logs", href: "/developer/logs" },
    ],
  },
  {
    name: "Settings",
    href: "/settings",
    icon: <Settings className="w-6 h-6" />,
    subItems: [
      {
        name: "SMTP",
        href: "/settings/smtp",
      },
      {
        name: "IMAP",
        href: "/settings/imap",
      },
      { name: "Account", href: "/settings" },
      { name: "Team", href: "/team" },
      { name: "Billing", href: "/settings/billing" },
    ],
  },
];

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const pathname = usePathname();
  const [expandedItems, setExpandedItems] = useState<Record<string, boolean>>({
    "/campaigns": false,
    "/automations": false,
    "/audience": true,
    "/analytics": true,
  });

  const { data: session } = useSession();
  const { apiFetch } = useApi();

  const toggleExpand = (href: string) => {
    setExpandedItems((prev) => ({
      ...prev,
      [href]: !prev[href],
    }));
  };

  const isActive = (href: string) => {
    return pathname === href || pathname.startsWith(href + "/");
  };

  const isSubItemActive = (href: string) => {
    return pathname === href;
  };

  useEffect(() => {
    const intercomRequest: any = {
      app_id: "ts43f4k1",
    };

    if (session?.user?.id) {
      intercomRequest.user_id = session?.user?.id ?? "";
      intercomRequest.name = session?.user?.name ?? "";
      intercomRequest.email = session?.user?.email ?? "";
      // @ts-ignore
      intercomRequest.created_at = session?.user?.createdAt ?? "";
      // @ts-ignore
      intercomRequest.user_hash = session?.user?.userHash ?? undefined;
    }

    Intercom(intercomRequest);
  }, [session?.user]);

  return (
    <div className="flex border-r border-muted h-full flex-col justify-between bg-sidebar w-auto lg:w-[250px]">
      <div>
        {/* Navigation */}
        <nav className="flex-1 pt-2">
          {navItems.map((item) => {
            const isItemActive = isActive(item.href);
            const isExpanded = expandedItems[item.href];
            const Icon = item.icon;

            return (
              <div key={item.href} className="mb-1">
                <div
                  className={cn(
                    "flex items-center px-4 py-1 text-sidebar-foreground hover:bg-accent cursor-pointer",
                    isItemActive &&
                      !item.subItems &&
                      "text-sidebar-foreground font-medium text-primary dark:text-red-500 dark:bg-black/40 bg-gray-100 border-l-2 border-primary "
                  )}
                  onClick={() => item.subItems && toggleExpand(item.href)}
                >
                  <Link
                    href={item.href}
                    className="flex items-center flex-1 gap-1"
                    onClick={(e) => item.subItems && e.preventDefault()}
                  >
                    {Icon}
                    <span className="text-sm">{item.name}</span>
                  </Link>
                  {item.subItems &&
                    (isExpanded ? (
                      <ChevronUp className="h-4 w-4 text-gray-500" />
                    ) : (
                      <ChevronDown className="h-4 w-4 text-gray-500" />
                    ))}
                </div>

                {/* Sub-items */}
                {item.subItems && isExpanded && (
                  <div className="ml-4 pl-4 border-l-2 border-gray-100">
                    {item.subItems.map((subItem) => {
                      const isSubActive = isSubItemActive(subItem.href);
                      return (
                        <Link
                          key={subItem.href}
                          href={subItem.href}
                          className={cn(
                            "flex font-geist items-center px-4 py-2 text-sm text-sidebar-foreground hover:bg-accent",
                            isSubActive &&
                              "text-primary bg-gray-100 border-l-2 border-primary -ml-[2px]"
                          )}
                        >
                          {subItem.name}
                        </Link>
                      );
                    })}
                  </div>
                )}
              </div>
            );
          })}
        </nav>
      </div>

      {/* Upgrade button */}
      <div className="p-4 mt-auto grid gap-2">
        <Sheet>
          <SheetTrigger asChild>
            <Button variant="outline" className="w-full font-medium">
              Compose
            </Button>
          </SheetTrigger>

          <SheetContent className="min-w-screen md:min-w-[800px] h-screen overflow-y-auto">
            <SheetHeader className="flex !mb-4 flex-col gap-2">
              <span className="text-2xl">Compose Email</span>
              <p className="text-sm">
                Write your email below to send it over to anyone
              </p>
            </SheetHeader>
            <ComposeEmail
              onSend={async (email) => {
                toast.promise(
                  async () => {
                    await apiFetch("emails", {
                      method: "POST",
                      body: JSON.stringify(email),
                    });
                  },
                  {
                    loading: "Sending email...",
                    success: "Email sent successfully",
                    error: "Failed to send email",
                  }
                );
              }}
            />
          </SheetContent>
        </Sheet>
        <Button variant="secondary" className="w-full font-medium">
          Upgrade
        </Button>
      </div>
    </div>
  );
}
