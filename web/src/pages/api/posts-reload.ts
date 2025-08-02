import type { APIRoute } from "astro";

export const GET: APIRoute = async ({ request }) => {
  const API_BASE =
    import.meta.env.PUBLIC_API_URL || "http://localhost:8080/api/v1";

  try {
    const response = await fetch(`${API_BASE}/posts`);

    if (!response.ok) {
      return new Response(JSON.stringify({ error: "Failed to fetch posts" }), {
        status: response.status,
        headers: { "Content-Type": "application/json" },
      });
    }

    const data = await response.json();

    return new Response(JSON.stringify(data.posts || []), {
      status: 200,
      headers: {
        "Content-Type": "application/json",
        "Cache-Control": "no-cache",
      },
    });
  } catch (error) {
    console.error("API route error:", error);
    return new Response(JSON.stringify({ error: "Internal server error" }), {
      status: 500,
      headers: { "Content-Type": "application/json" },
    });
  }
};
