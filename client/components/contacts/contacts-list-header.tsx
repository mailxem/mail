"use client";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Download, Plus } from "lucide-react";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetDescription,
  SheetTrigger,
} from "@/components/ui/sheet";
import { useState } from "react";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { SheetFooter } from "@/components/ui/sheet";
import { toast } from "sonner";
import { useApi } from "@/hooks/use-api";
import { useTeam } from "@/app/providers/team-provider";
import { useMailingLists } from "@/app/providers/mailinglist-provider";

export function ContactsListHeader() {
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [newList, setNewList] = useState({ name: "", description: "" });
  const [isLoading, setIsLoading] = useState(false);
  const { apiFetch } = useApi();
  const { team } = useTeam();
  const { refetch } = useMailingLists();

  const createList = async () => {
    if (!newList.name.trim() || isLoading) return;
    setIsLoading(true);
    try {
      const response = await apiFetch("mailing-lists", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ ...newList, teamId: team?.id }),
      });

      if (!response.ok) throw new Error("Failed to create list");

      await refetch();
      setNewList({ name: "", description: "" });
      setIsDialogOpen(false);
      toast.success("Contact list created successfully");
    } catch (error) {
      toast.error("Failed to create contact list");
    } finally { setIsLoading(false); }
  };

  return (
    <div className="flex items-center gap-3">
      <Sheet open={isDialogOpen} onOpenChange={setIsDialogOpen}>
        <SheetTrigger asChild>
          <Button className="text-white">
            <Plus className="h-4 w-4 mr-2" />
            New List
          </Button>
        </SheetTrigger>
        <SheetContent className="sm:max-w-[500px]">
          <SheetHeader className="space-y-4 pb-6 border-b">
            <SheetTitle className="text-2xl font-medium text-[#241C15]">
              Create New Contact List
            </SheetTitle>
            <SheetDescription className="text-base text-gray-600">
              Create a new list to organize your contacts and manage your
              audience more effectively.
            </SheetDescription>
          </SheetHeader>
          <div className="space-y-6 py-6">
            <div className="space-y-2">
              <label htmlFor="name" className="text-sm font-medium ">
                List Name
              </label>
              <Input
                id="name"
                value={newList.name}
                onChange={(e) =>
                  setNewList({ ...newList, name: e.target.value })
                }
                placeholder="e.g., Newsletter Subscribers"
                className="border-gray-200"
              />
            </div>
            <div className="space-y-2">
              <label htmlFor="description" className="text-sm font-medium ">
                Description
              </label>
              <Textarea
                id="description"
                value={newList.description}
                onChange={(e) =>
                  setNewList({ ...newList, description: e.target.value })
                }
                placeholder="Add a description to help you remember what this list is for..."
                className="min-h-[100px] border-gray-200"
              />
            </div>
          </div>
          <SheetFooter className="pt-4 border-t">
            <Button
              variant="outline"
              onClick={() => setIsDialogOpen(false)}
              disabled={isLoading}
              className="border-gray-200"
            >
              Cancel
            </Button>
            <Button
              onClick={createList}
              disabled={isLoading}
              className="bg-[#007C89] text-white hover:bg-[#005F6B]"
            >
              Create List
            </Button>
          </SheetFooter>
        </SheetContent>
      </Sheet>
    </div>
  );
}
