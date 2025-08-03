const API_BASE =
  import.meta.env.PUBLIC_API_URL || "http://localhost:8080/api/v1";

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

  login: (credentials: any) =>
    apiCall("/auth/login", { method: "POST", body: credentials }),

  health: () => apiCall("/health", { method: "GET" }),
};
