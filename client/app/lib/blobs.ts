import { put } from "@vercel/blob";

export async function uploadFile(file: any) {
    const blob = await put(file.name, file, {
        access: "public",
    });
    return blob.url;
}