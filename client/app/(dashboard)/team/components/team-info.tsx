"use client";

import { useTeam } from "@/app/providers/team-provider";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { MoreHorizontal, Shield, User } from "lucide-react";

interface TeamMember {
  id: string;
  email: string;
  firstName: string | null;
  lastName: string | null;
  role: "ADMIN" | "USER";
}

export function TeamInfo() {
  const { team } = useTeam();

  if (!team?.users?.length) {
    return (
      <div className="flex flex-col items-center justify-center p-8 text-center">
        <User className="h-10 w-10 text-muted-foreground mb-4" />
        <p className="text-lg font-medium">No team members found</p>
        <p className="text-sm text-muted-foreground">
          Start by inviting team members to collaborate
        </p>
      </div>
    );
  }

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Member</TableHead>
          <TableHead>Role</TableHead>
          <TableHead>Email</TableHead>
          <TableHead className="w-[100px]">Actions</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {team.users.map((member: any) => (
          <TableRow key={member.id}>
            <TableCell>
              <div className="flex items-center gap-3">
                <Avatar className="h-8 w-8 bg-secondary">
                  <AvatarImage
                    src={`https://api.dicebear.com/9.x/lorelei/svg?seed=${member.firstName}`}
                    alt={member.firstName || member.email}
                  />
                  <AvatarFallback>
                    {(member.firstName || member.email).charAt(0).toUpperCase()}
                  </AvatarFallback>
                </Avatar>
                <div>
                  <p className="font-medium">
                    {member.firstName || "Unnamed"}
                  </p>
                  <p className="text-sm text-muted-foreground">
                    {member.email}
                  </p>
                </div>
              </div>
            </TableCell>
            <TableCell>
              <Badge
                variant={member.role === "ADMIN" ? "default" : "secondary"}
                className="flex w-fit items-center gap-1"
              >
                {member.role === "ADMIN" && <Shield className="h-3 w-3" />}
                {member.role}
              </Badge>
            </TableCell>
            <TableCell className="font-mono text-sm">{member.email}</TableCell>
            <TableCell>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="ghost" size="icon">
                    <MoreHorizontal className="h-4 w-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuSub>
                    <DropdownMenuSubTrigger>Change Role</DropdownMenuSubTrigger>
                    <DropdownMenuSubContent>
                      <DropdownMenuItem onClick={() => {}}>
                        Admin
                      </DropdownMenuItem>
                      <DropdownMenuItem onClick={() => {}}>
                        Member
                      </DropdownMenuItem>
                      <DropdownMenuItem onClick={() => {}}>
                        Super Admin
                      </DropdownMenuItem>
                    </DropdownMenuSubContent>
                  </DropdownMenuSub>
                  <DropdownMenuItem
                    className="text-red-600"
                    onClick={() => {
                      // Handle member removal
                    }}
                  >
                    Remove Member
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}
