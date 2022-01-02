# Création d'une blockchain pour le suivi des numéros de série
### Cas de lutte contre la contrefaçon des médicaments


# Setup: 
- Sur Linux, installer go dans /usr/local et déplacer le projet dans /usr/local/src/blockchain_go
```bash
make setup
make build
```

# Scenario:
### 1. Lancer le noeud "server/tracker". 
Tous les clients peuvent ajouter et faire la mise à jour. Ce nœud n'est un serveur que dans le sens où il prend les nouveaux blocs des miners et les diffuse aux nœuds clients. Il est toujours le miner le plus rapide à gagner. Le mécanisme de concensus n'est pas entièrement décentralisé pour le moment.

- ./server.sh localhost 3000 localhost 2000

### 2. Lancer le miner. 
Nous avons configuré le miner pour qu'il commence son travaille après avoir reçu deux transactions. Dans une nouvelle fenêtre, tapez :

- ./miner.sh localhost 2999

### 3. Démarrez le client 1 et créez deux numéros de série. Dans une nouvelle fenêtre, tapez :

- ./wallet.sh localhost 3001

Dans le CLI, tapez :
- add [coller l'addresse du noueud depuis stdout] 3001 0 (3001 est le numéro de série ; 0 est un espace réservé pour l'ancien champ "salt". Nous pourrons utiliser ce champ pour renforcer la confidentialité, mais gardons le placeholder pour le moment.)
- add [coller l'addresse du noueud depuis stdout] 30013001 0

### 4. Démarrez le client deux, créez un numéro de série et transférez un numéro de série existant. Dans une nouvelle fenêtre, tapez :

- ./wallet.sh localhost 3002

Dans le CLI, tapez :
- add [coller l'addresse du noeud depuis stdout] 3002 0

Retour à la fenêtre du client 1, tapez :
- send [address du noeud 1] [addresse du noeud 2] 30013001 0  (transferrer '30013001' depuis le premier noeud au deuxième)

### 5. Démarrez le client 3, créez un numéro de série et transférez à nouveau le numéro de série '30013001'. Dans une nouvelle fenêtre, tapez :

- ./wallet.sh localhost 3003

Dans le CLI, tapez :
- add [coller l'addresse du noeud depuis stdout] 3003 0
- send [address du noeud 2] [address du noeud 3] 30013001 0  (transferrer '30013001' depuis le premier noeud au deuxième)


### Autres choses à essayer dans le CLI :
- get 30013001 0  (get [numero de série] [salt])
- print
