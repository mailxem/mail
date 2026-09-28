import { Plus } from "lucide-react";
import { Button } from "../ui/button";
import {
  Sheet,
  SheetTrigger,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetDescription,
  SheetFooter,
} from "../ui/sheet";
import { useState } from "react";
import { Input } from "../ui/input";

export function TagsPageHeader() {
  const [isCreateTagOpen, setIsCreateTagOpen] = useState(false);
  const [newTag, setNewTag] = useState("");
  return (
    <div className="flex gap-2">
      <Button variant="outline" className="border-gray-200">
        Bulk tag
      </Button>
      <Sheet open={isCreateTagOpen} onOpenChange={setIsCreateTagOpen}>
        <SheetTrigger asChild>
          <Button>
            <Plus className="h-4 w-4 mr-2" />
            Create new tag
          </Button>
        </SheetTrigger>
        <SheetContent>
          <SheetHeader>
            <SheetTitle>Create New Tag</SheetTitle>
            <SheetDescription>
              Add a new tag to help organize your contacts
            </SheetDescription>
          </SheetHeader>
          <div className="py-6">
            <label className="text-sm font-medium">Tag Name</label>
            <Input
              value={newTag}
              onChange={(e) => setNewTag(e.target.value)}
              placeholder="Enter tag name"
              className="mt-2"
            />
          </div>
          <SheetFooter>
            <Button variant="outline" onClick={() => setIsCreateTagOpen(false)}>
              Cancel
            </Button>
            <Button
              onClick={() => {
                // Handle tag creation
                setIsCreateTagOpen(false);
              }}
              className="bg-[#007C89] text-white hover:bg-[#005F6B]"
            >
              Create Tag
            </Button>
          </SheetFooter>
        </SheetContent>
      </Sheet>
    </div>
  );
}
