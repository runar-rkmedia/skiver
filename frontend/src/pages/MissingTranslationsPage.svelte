<script lang="ts">
  import EntityList from '../components/EntityList.svelte'
  import ListItem from '../components/ListItem.svelte'
  import { db } from '../api'
  import { state } from '../state'
  import formatDate from 'dates'
import MissingTranslations from 'components/MissingTranslations.svelte'
</script>


<h2>Missing Translations</h2>

<MissingTranslations/>
