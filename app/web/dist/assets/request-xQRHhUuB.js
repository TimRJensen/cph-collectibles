async function n(r,e,s=null){try{return(await fetch("https://localhost"+r,{method:e,body:s})).json()}catch(t){return t}}export{n as r};
