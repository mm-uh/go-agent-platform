# Plataforma de Agentes

Este proyecto representa una Plataforma de Agentes, como respuesta a un ejercicio final de clases de la asignatura `Sistemas Distribuidos` del plan de clases de la Universidad de la Habana, correspondiente a los estudiantes de 4to año de `Ciencias de la Computación`.

Este es el proyecto principal de la organizacion [mm-uh](https://github.com/mm-uh). Esta consta de todas las herramientas usadas para la construcción de este proyecto, además de varias librerias para el uso de esta plataforma desde varios lenguajes de programación([Python], Golang), además de constar con una forma de generar librerias clientes que consumen los servicios de esta plataforma de una manera segura y rápida para otro conjunto de lenguajes de programacion, gracias a [Swagger](https://swagger.io).

## Instalando

Para ejecutar un nodo de nuestra plataforma necesita tener instalado [golang](https://golang.org). No necesita ningun componente adicional para correr nuestro entorno y tener una plataforma perfectamente funcional.

```sh
go get -t -v github.com/mm-uh/go-agent-platform
```

### Usando docker

En caso de usted no querer instalar `Golang` para correr la plataforma, puede usar alternativamente [docker](https://www.docker.com).

En tal caso, de tener instalado correctamente docker, solo necesita o bien construirse su propia imagen de docker y correrla, o descargarsela desde [github](https://github.com/mm-uh/go-agent-platform/releases)

En caso de usted desear construir localmente la imagen, solo tiene que correr el siguiente comando:

```sh
docker build -t <image-name> .
```

## Ejecutando nodos

Tenemos dos variantes directas para la ejecucion de los nodos, una compilando directamente nuestra aplicacion, y la otra usando docker.

### Ejecucion del compilado

Para ejecutar la aplicacion solo se necesita anteriormente ejecutar la siguiente linea:

```sh
make build
```

Este comando se encarga de compilar la apicacion y generar un binario por defecto `platform`.

Si usted desea cambiarle el nombre solo necesita correlo de esta forma:

```sh
DEFAULT_APP_NAME=<selected name> make build
```

### Flags de ejecucion

```sh
❯ ./platform --help
Invalid number arguments
Usage:
app <ip_to_run> <port_to_run> [remote_ip] [remote_port]

<ip_to_run>   -> IP that the app will take
<host_to_run> -> Port that the app will run
[remote_ip]   -> (Optional) IP of another platform
[remote_port] -> (Optional) Port of another platform
```

### En el caso de docker

Posteriormente de tener la imagen, solo necesita correr la imagen con los parametros de la aplicacion, ejemplo:

```sh
docker run --network=host -it --rm -p [bind ports propertly...] <image-name> [normal flags]
```

Ejemplo:

```sh
docker run --network=host -it --rm -p 8082:8082 -p 9082:9082 -p 10082:10082 mm-uh/go-agent-platform:latest 192.168.0.110 8082 192.168.0.110 8081
```

## Adicionar agentes a la plataforma

Para la adicion de agentes a la plataforma, se recomienda el uso de nuestras librerias:

- [agent-libpython](http://github.com/mm-uh/agent-libpython)
- [agent-libgo](http://github.com/mm-uh/agent-libgo)

Para la generacion la genaracion de las librerias que se usan para consumir los servicios del API REST que tiene cada nodo de la plataforma usamos swagger, para describir dicha API y posteriormente se generaron las librerias.

Ambas librerias contienen ademas una estructura o clase que recubre esos pedidos que facilitan la vida del programador y hace varias cosas, como es por ejemplo, implementar desde el primer momento, y ageno al usuario, un sistema de tolerancia-falla intrinseco en los puntos de acceso de la plataforma.

Para mas detalles de como generar librerias de otros lenguajes ver [openapi](https://openapi-generator.tech/).

## Arquitectura

Nuestra plataforma esta basada en un conjunto de nodos, donde dos nodos estan conectados si estan sobre la misma base de datos. Los nodos de nuestra plataforma utilizan como base de datos una tabla de hash distribuida(conocidas en la literatura como Distributed Hash Table[DHT]), basada en el algoritmo de `Kademlia`.

Asociados a esta plataforma, se inscriben agentes, y la informacion de contacto se guarda por la plataforma par su posterior uso.

### Interaccion con la plataforma

Para la coneccion con la plataforma usamos el protocolo de coneccion REST: el API se encuentra totalmente descrita [aqui](https://github.com/mm-uh/go-agent-platform/blob/develop/swagger.yml)

- /getAgent/{Name}
  - Obtener un agente a partir de un nombre

- /getPeers
  - Obtiene una secuencia de los k nodos mas cercanos al nodo que realiza la peticxion

- /registerAgent
  - A partir de una definicion de agente(Se genera de forma automatica si se usa swagger), se registra la estancia de el agente

- /addEndpoints
  - Adiciona a un agente ya registrado un conjunto de puntos de acceso nuevos

- /recoverAgent
  - Recupera mediante la contrasena un agente

- /editAgent
  - Modifica las propiedades de un agente ya registrado

- /getAllAgentsNames
  - Obtiene todos los nombres de los agentes registrados

- /getAllFunctionsNames
  - Obtiene todos los nombres de las funciones registradas

- /getAgentsForFunction/{Name}
  - Obtiene el nombre de todos los agentes que pertenecen a una funcion dada

- /getSimilarAgents/{Name}
  - Obtiene la localizacion de un agente que dado un criterio determinado de tolerancia, establece que se pueda enviar un agente que sustituya a el que se envia

Todos los puntos de acceso son relativos a `/api/v1`, ejemplo : `http://localhost:10080/api/v1/getPeers`

### Publicar servicios

Para publicar un Agente se usa el punto de acceso `/registerAgent` del API y en el caso de las librerias, su correspondiente encapzulamiento. A partir de ese momento, con solo tener los servicios tanto de comprobacion, como de documentacion, como del servicio en si,se puede comenzar a usar el agente registrado. 

### Buscar sevicios

Para buscar servicios en nuestra plataforma, se establecen dos posibles escenarios.

Uno en el que usted necesite de un agente en especifico(`/getAgent/{Name}`), el cual retorna la forma de contactar un nodo que se encuetre funcionando.

El otro caso puede ser que se necesite las buscar por una funcionalidad, en cuyo caso solo se debe hacer un pedido al punto de acceso `getAgentsForFunction/{Name}` y se obtienen todos los agentes que tienen dicha funcionalidad.
