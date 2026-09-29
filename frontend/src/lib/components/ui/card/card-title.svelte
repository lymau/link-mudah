<script>
	import { cn } from "$lib/utils.js";
	let {
		ref = $bindable(null),
		class: className,
		children,
		tag: Tag = "h3",
		...restProps
	} = $props();
</script>

<svelte:element
	this={Tag}
	bind:this={ref}
	data-slot="card-title"
	class={cn("font-semibold leading-none tracking-tight", className)}
	{...restProps}
>
	{@render children?.()}
</svelte:element>