<script>
export default {
	props: ['msg']
}
</script>

<template>
	<div class="error-msg" role="alert" aria-live="polite">
		{{ msg }}
	</div>
</template>

<style scoped>
.error-msg {
	margin: .5rem 0 1rem;
  padding: .75rem 1rem;
  border-radius: var(--radius);
  background: color-mix(in oklab, var(--bg) 80%, #ef4444 20%);
  border: 1px solid color-mix(in oklab, #ef4444 60%, var(--border) 40%);
  color: #ffe6e6;
  box-shadow: 0 0 10px rgba(239,68,68,.25);
  font-size: .95rem;
}
</style>
