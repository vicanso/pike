import { h, app } from 'hyperapp';

const Directors = ({ state }) => {
  if (!state.directors) {
    return <div class="directorsWrapper container">
      <p class="tac">Loading...</p>
    </div>
  }

  const arr = state.directors.map((director) => {
    const getItems = (director) => {
      const keys = [
        'hosts',
        'passes',
        'prefixs',
        'backends',
        'policy',
      ];
      const format = (data) => {
        if (Array.isArray(data)) {
          return <ul>
            {
              data.map((item) => <li>
                {item}
              </li>)
            }
          </ul>;
        }
        return data;
      };
      return keys.map((key) => {
        let value = director[key];
        if (key === 'backends') {
          value = value.map((item) => {
            let found = null;
            director.upstream.hosts.forEach((upstream) => {
              if (upstream.host === item) {
                found = upstream;
              }
            });
            if (!found) {
              return item;
            }
            const status = found.disabled ? "disabled" : "enabled";
            return <span>
              {item}({status})
              {
                found.healthy != 0 &&
                <span class="mleft5 greenColor">healthy</span>
              }
              {
                found.healthy == 0 &&
                <span class="mleft5 redColor">sick</span>
              }
            </span>
          });
        }
        return <tr>
          <td class="name">{key}</td>
          <td>{format(value)}</td>
        </tr>
      });
    }
    return <div class="directorWrapper">
      <h5>{director.name}</h5>
      <table class="table">
        <thead><tr>
          <th class="name">Name</th>
          <th>Setting</th>
        </tr></thead>
        <tbody>
          {getItems(director)}
        </tbody>
      </table>
    </div>
  });
  return <div class="directorsWrapper container">
    {arr}
  </div>
}

export default Directors
