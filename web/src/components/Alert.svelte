<script>
  export let message;
  import { fade,fly } from 'svelte/transition';
  import { quintOut } from 'svelte/easing';

  setTimeout(function() {
    message = null;
  }, 6000);
</script>

<style>
  .alert {
    color: #fff;
    display: inline-block;
    font-size: 32px;
    margin: 0 auto;
    position: absolute;
    top: 50%;
  }

  .alert-container {
    height: 100%;
    width: 100%;
  }
</style>

{#if message}
  <div class="alert-container">
    <div class="alert" in:fly="{{delay: 250, duration: 3000, easing: quintOut}}" out:fade>
      {message}
    </div>
  </div>
{/if}
