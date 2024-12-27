Para limpar completamente o Docker, removendo todos os contêineres, imagens, volumes e redes, você pode usar os seguintes comandos. **Atenção**: Esses comandos irão remover todos os dados do Docker, então certifique-se de que você realmente deseja fazer isso antes de prosseguir.

1. **Parar todos os contêineres em execução**:
   ```bash
   docker stop $(docker ps -aq)
   ```

2. **Remover todos os contêineres**:
   ```bash
   docker rm $(docker ps -aq)
   ```

3. **Remover todas as imagens**:
   ```bash
   docker rmi $(docker images -q)
   ```

4. **Remover todos os volumes**:
   ```bash
   docker volume rm $(docker volume ls -q)
   ```

5. **Remover todas as redes não utilizadas**:
   ```bash
   docker network prune -f
   ```

6. **Remover todos os dados do sistema Docker**:
   ```bash
   docker system prune -a --volumes -f
   ```

### Explicação:

- **`docker stop $(docker ps -aq)`**: Para todos os contêineres em execução.
- **`docker rm $(docker ps -aq)`**: Remove todos os contêineres, parados ou em execução.
- **`docker rmi $(docker images -q)`**: Remove todas as imagens.
- **`docker volume rm $(docker volume ls -q)`**: Remove todos os volumes.
- **`docker network prune -f`**: Remove todas as redes não utilizadas.
- **`docker system prune -a --volumes -f`**: Remove todos os dados do sistema Docker, incluindo contêineres parados, imagens não utilizadas, redes e volumes.

Esses comandos irão limpar completamente o seu ambiente Docker. Use-os com cuidado, pois não há como desfazer essas ações.
