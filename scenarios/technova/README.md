# Galah 1.1.1 — templates TechNova

Implémenté dans `/home/pa/galah` sur `pa@192.168.64.7:64295`.

## Résultat

- Configuration optionnelle `scenario` : `enabled`, `name`, `templates_dir`, `routes` avec `pattern`, `template`, `fields`.
- Templates chargés au démarrage, expressions régulières compilées, première route correspondante retenue. Le matching utilise `URL.Path`, sans query string.
- Le LLM des routes mappées fournit exactement les champs demandés, sous forme de chaînes de texte. Les prompts HTTP complets restent utilisés uniquement pour les routes non mappées.
- Rendu via `html/template`, échappement contextuel, refus des champs manquants et des templates hors du répertoire configuré.
- Même `llm.JSONResponse`, même stockage et consultation du cache, même serveur et journalisation. Les nouvelles réponses restent journalisées avec la source existante `llm`.
- Une erreur de génération ou de rendu sur une route mappée produit une erreur et n'est pas mise en cache ; elle ne déclenche pas une deuxième génération complète.
- `profile.yaml` est descriptif ; Galah charge `config.yaml`, pas le profil automatiquement.
- Scénario exemple : `/`, `/it-support` et suffixes, `/projects` et suffixes, `/kb` et suffixes, `/about` et sous-chemins. Les autres chemins conservent le fallback complet.
- Dockerfile `think=false` et serveur `WriteTimeout=120s` conservés.

## Validation effectuée

`go test ./...` passe, y compris dans l'image reconstruite. Tests ajoutés pour l'échappement, les erreurs de templates/routes, les réponses LLM invalides, le pipeline, le cache et le fallback. Les cinq templates et la configuration de déploiement préparée ont été validés (quatre ports, cinq routes).

Image reconstruite : `galah-local:1.1.1`.
Identifiant : `sha256:05a828aac870c2e82ac1830ce7aca883ddab29b6ff21486eb3ec5a201d82309a`.
Ancienne image conservée : `galah-local:1.1.1-before-scenarios-20261003`.

Le service actif n'a pas été recréé. Aucun test avec le LLM réel n'a été effectué.

## Déploiement — depuis la session SSH T-Pot

La configuration actuelle contient des lignes à un seul espace dans le bloc `system_prompt` (notamment les URL canoniques), alors que le bloc exige au moins deux espaces. Le conteneur actuel redémarre avec une erreur YAML. Le fichier préparé `data/galah/config/config.technova.yaml` corrige cette indentation et ajoute `scenario`, en conservant le texte des prompts, les ports et les profils TLS. Le fichier actif est resté intact.

Relire la différence avant de déployer :

```bash
cd ~/tpotce
diff -u data/galah/config/config.yaml data/galah/config/config.technova.yaml
```

Puis :

```bash
cd ~/tpotce
cp -p data/galah/config/config.yaml "data/galah/config/config.yaml.bak.$(date +%Y%m%d-%H%M%S)"
mkdir -p data/galah/config/scenarios
cp -a ~/galah/scenarios/technova data/galah/config/scenarios/
cp data/galah/config/config.technova.yaml data/galah/config/config.yaml
docker compose config --quiet
docker compose up -d --no-deps --force-recreate galah
docker compose ps galah
docker compose logs --tail=60 galah
```

Le répertoire de templates est couvert par le montage existant `/galah/config` : aucun changement Compose n'est requis. `--no-deps` limite l'opération au service Galah ; ne pas exécuter `docker compose down`.

Vérifier une route mappée et le cache avec la même URL :

```bash
url="http://127.0.0.1/it-support-$(date +%s)"
curl --max-time 120 -sS -D /tmp/galah-headers.txt "$url" -o /tmp/galah-ticket.html
cat /tmp/galah-headers.txt
head -c 1000 /tmp/galah-ticket.html
curl --max-time 120 -sS "$url" -o /tmp/galah-ticket-cached.html
cmp /tmp/galah-ticket.html /tmp/galah-ticket-cached.html
docker compose logs --tail=40 galah
```

Les logs doivent montrer `source: llm` pour la génération puis `source: cache` pour la répétition, si le cache est activé. Un corps identique seul ne prouve pas le cache : consulter les logs.

Vérifier le fallback :

```bash
curl --max-time 120 -i "http://127.0.0.1/unmapped-check-$(date +%s)"
```

## Points de vigilance et alternatives

- `rules.yaml` est inchangé. Sa règle active `example default response`, regex `^/$`, type `static`, template `templates/default.json`, court-circuite le scénario pour `/`. Les tests doivent utiliser `/it-support-...`. La page d'accueil TechNova ne sera utilisée qu'après désactivation explicite de cette règle.
- Le cache existant est conservé : une URL déjà mise en cache peut continuer à servir son ancien HTML. Utiliser des suffixes nouveaux pour vérifier la génération. Les templates sont chargés au démarrage : leurs modifications nécessitent la recréation ou le redémarrage de Galah.
- Pour revenir au pipeline complet avec la nouvelle image, passer `scenario.enabled` à `false` dans le fichier YAML validé puis recréer uniquement Galah. Avantage : retour simple au fonctionnement original ; inconvénient : générations HTML complètes plus longues. Ne pas restaurer sans correction la sauvegarde YAML invalide.
- Pour mapper davantage d'URL, ajouter des routes avant les éventuelles routes génériques. Une route `.*` supprimerait le fallback pour tous les chemins.
- Les champs sont volontairement des chaînes de texte ; objets imbriqués, nombres et valeurs nulles sont refusés. Cela simplifie les templates et la validation. Aucune modification des paramètres du modèle n'a été faite.

Sources techniques : [html/template (Go)](https://pkg.go.dev/html/template), [docker compose up](https://docs.docker.com/reference/cli/docker/compose/up/).
