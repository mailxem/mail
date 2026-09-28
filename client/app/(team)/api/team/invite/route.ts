import { auth } from "@/auth";
import { APIService } from "@/lib/services/api";
import { NextResponse } from "next/server";

export async function POST(request: Request): Promise<NextResponse> {
  try {
    const session = await auth();
    if (!session?.user) {
      return NextResponse.json(
        { error: "Unauthorized" },
        { status: 401 }
      );
    }

    const body = await request.json();
    const { email, name, role } = body;

    // Validate input
    if (!email || !name || !role) {
      return NextResponse.json(
        { error: "Email, name and role are required" },
        { status: 400 }
      );
    }
    
    // Use APIService for the registration
    await new APIService("auth", session).post("invite", {
      email: email,
      name: name,
      role: role,
    });

    return NextResponse.json(
      { message: "Invitation sent successfully" },
      { status: 200 }
    );
  } catch (error) {
    console.error("Error sending invitation:", error);
    return NextResponse.json(
      { error: "Failed to send invitation" },
      { status: 500 }
    );
  }
}
