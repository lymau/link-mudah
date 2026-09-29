<script module>
	import { tv } from "tailwind-variants";
	import { cn } from "$lib/utils.js";

	export const buttonVariants = tv({
		base: "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-all focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg:not([class*='size-'])]:size-4 shrink-0 select-none cursor-pointer active:scale-[0.98]",
		variants: {
			variant: {
				default: "bg-primary text-primary-foreground shadow-sm hover:bg-primary/90",
				destructive: "bg-destructive text-destructive-foreground shadow-xs hover:bg-destructive/90 focus-visible:ring-destructive/20",
				outline: "border border-input bg-background shadow-xs hover:bg-accent hover:text-accent-foreground",
				secondary: "bg-secondary text-secondary-foreground shadow-xs hover:bg-secondary/80",
				ghost: "hover:bg-accent hover:text-accent-foreground",
				link: "text-primary underline-offset-4 hover:underline",
			},
			size: {
				default: "h-9 px-4 py-2",
				xs: "h-7 rounded px-2 text-xs",
				sm: "h-8 rounded-md px-3 text-xs",
				lg: "h-10 rounded-md px-6 text-sm",
				xl: "h-11 rounded-md px-8 text-base",
				icon: "size-9",
				"icon-xs": "size-7",
				"icon-sm": "size-8",
				"icon-lg": "size-10",
			},
		},
		defaultVariants: {
			variant: "default",
			size: "default",
		},
	});
</script>

<script>
	let {
		class: className,
		variant = "default",
		size = "default",
		ref = $bindable(null),
		href = undefined,
		type = "button",
		disabled,
		children,
		...restProps
	} = $props();
</script>

{#if href}
	<a
		bind:this={ref}
		data-slot="button"
		class={cn(buttonVariants({ variant, size }), className)}
		href={disabled ? undefined : href}
		aria-disabled={disabled}
		role={disabled ? "link" : undefined}
		tabindex={disabled ? -1 : undefined}
		{...restProps}
	>
		{@render children?.()}
	</a>
{:else}
	<button
		bind:this={ref}
		data-slot="button"
		class={cn(buttonVariants({ variant, size }), className)}
		{type}
		{disabled}
		{...restProps}
	>
		{@render children?.()}
	</button>
{/if}