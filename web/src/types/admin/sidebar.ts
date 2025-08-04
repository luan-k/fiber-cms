// Icon Path type
export type IconPath = `${Lowercase<string>}` extends infer T
? T extends `${string}.${string}` // Prevent file extensions
    ? never
    : T extends `${string}` // Basic string check
    ? T extends `${infer C}` // Recursive character check
        ? C extends Lowercase<'abcdefghijklmnopqrstuvwxyz0123456789-_'> ? T : never // Only allow letters and the "-" and "_" characters
        : never
    : never
: never;

export interface Section {
    icon: IconPath, // Shouldn't include a file extension, and it exclusively represents files in assets/icons
    name: string,
    link: string
}

export type Navigation = Section[][]