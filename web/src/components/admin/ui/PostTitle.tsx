import type { JSX } from 'react';
import type { Post } from "@/lib/types";

export default function PostTitle({ value }: { value: Post }): JSX.Element {
    return <h2 className="gl-post__title">{value.title}</h2>;
}