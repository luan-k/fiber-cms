const API_BASE =
  typeof window === "undefined"
    ? "http://api:8080/api/v1"
    : import.meta.env.PUBLIC_API_URL || "http://localhost:8080/api/v1";

// debugging logs for envs
console.log(
  "API_BASE:",
  API_BASE,
  "Context:",
  typeof window === "undefined" ? "server" : "browser"
);

console.log("Environment debug:", {
  PUBLIC_API_URL: import.meta.env.PUBLIC_API_URL,
  SERVER_API_URL: import.meta.env.SERVER_API_URL,
  NODE_ENV: import.meta.env.NODE_ENV,
  MODE: import.meta.env.MODE,
  isWindow: typeof window !== "undefined",
  allEnv: import.meta.env,
});

console.log("API_BASE:", API_BASE);

interface ApiOptions {
  token?: string;
  method?: string;
  body?: any;
}

interface ApiResponse<T> {
  meta?: {
    count: number;
    limit: number;
    offset: number;
    total: number;
  };
  posts?: T[];
  taxonomies?: T[];
  media?: T[];
  users?: T[];
}

export async function apiCall(endpoint: string, options: ApiOptions = {}) {
  const { token, method = "GET", body } = options;

  const config: RequestInit = {
    method,
    headers: {
      "Content-Type": "application/json",
      ...(token && { Authorization: `Bearer ${token}` }),
    },
  };

  if (body) {
    config.body = JSON.stringify(body);
  }

  try {
    const url = `${API_BASE}${endpoint}`;
    console.log("Making API call to:", url);

    const response = await fetch(url, config);

    if (response.status === 401 && typeof window !== "undefined") {
      const { authManager } = await import("./auth.ts");
      const refreshed = await authManager.refreshAccessToken();

      if (refreshed) {
        const newToken = authManager.getAccessToken();
        if (newToken) {
          config.headers = {
            ...config.headers,
            Authorization: `Bearer ${newToken}`,
          };
          const retryResponse = await fetch(url, config);
          if (retryResponse.ok) {
            return retryResponse.json();
          }
        }
      }

      window.location.href = "/login";
      throw new Error("Authentication required");
    }

    if (!response.ok) {
      throw new Error(`API Error: ${response.status} ${response.statusText}`);
    }

    return response.json();
  } catch (error) {
    console.error("API call failed:", error);
    throw error;
  }
}

export const api = {
  login: (credentials: { username: string; password: string }) =>
    apiCall("/auth/login", { method: "POST", body: credentials }),

  renewAccessToken: (data: { refresh_token: string }) =>
    apiCall("/auth/refresh", { method: "POST", body: data }),

  logout: (data: { refresh_token: string }) =>
    apiCall("/auth/logout", { method: "POST", body: data }),
  getPosts: async () => {
    const response: ApiResponse<any> = await apiCall("/posts");
    return {
      data: response.posts || [],
      meta: response.meta || { count: 0, limit: 10, offset: 0, total: 0 },
    };
  },
  getPost: (id: string) => apiCall(`/posts/${id}`),
  getPostsByUser: (userId: string) => apiCall(`/posts/user/${userId}`),

  getUsers: async (token?: string) => {
    const response: ApiResponse<any> = await apiCall("/users", { token });
    return {
      data: response.users || [],
      meta: response.meta || { count: 0, limit: 10, offset: 0, total: 0 },
    };
  },
  getUser: (id: string) => apiCall(`/users/${id}`),
  getUserByUsername: (username: string) =>
    apiCall(`/users/username/${username}`),

  getTaxonomies: async () => {
    const response: ApiResponse<any> = await apiCall("/taxonomies");
    return {
      data: response.taxonomies || [],
      meta: response.meta || { count: 0, limit: 10, offset: 0, total: 0 },
    };
  },
  getTaxonomy: (id: string) => apiCall(`/taxonomies/${id}`),
  getPopularTaxonomies: async () => {
    const response: ApiResponse<any> = await apiCall("/taxonomies/popular");
    return {
      data: response.taxonomies || [],
      meta: response.meta || { count: 0, limit: 10, offset: 0, total: 0 },
    };
  },
  searchTaxonomies: (query: string) => apiCall(`/taxonomies/search?q=${query}`),

  getMedia: async (token?: string) => {
    const response: ApiResponse<any> = await apiCall("/media", { token });
    return {
      data: response.media || [],
      meta: response.meta || { count: 0, limit: 10, offset: 0, total: 0 },
    };
  },
  getMediaById: (id: string) => apiCall(`/media/${id}`),
  getPopularMedia: async () => {
    const response: ApiResponse<any> = await apiCall("/media/popular");
    return {
      data: response.media || [],
      meta: response.meta || { count: 0, limit: 10, offset: 0, total: 0 },
    };
  },
  searchMedia: (query: string) => apiCall(`/media/search?q=${query}`),

  health: () => apiCall("/health", { method: "GET" }),
};
