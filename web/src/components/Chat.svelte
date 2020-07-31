<script>
    import { afterUpdate } from 'svelte';
    export let messages;

    afterUpdate(() => {
      var chat = document.querySelector('#chat');
      chat.scrollTo(0, chat.clientHeight);
    });
</script>
<style>

	table {
    border: none;
    color: #fff;
  }

  td {
    padding: 3px 5px;
  }

  td.username {
    font-weight: bold;
    text-align: right;
  }

  td.username::after {
    content: ': ';
  }

  #chat {
    background-color: rgba(0,0,0, 0.5);
    font-family: 'Roboto', sans-serif;
    font-weight: 100;
    height: 565px;
    overflow: hidden;
    padding: 10px;
    position: fixed;
    right: 10px;
    top: 76px;
    width: 365px;
  }
</style>
<div id="chat">
    <table>
        {#each messages as msg}
          <tr>
            <td class="username" style="color: {msg.user.color};">{msg.user.displayName}</td>
            <td class="message">{msg.text}</td>
          </tr>
        {/each}
    </table>
</div>