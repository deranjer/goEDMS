import React, { useState, useCallback } from "react";
import Tree, { TreeNode} from "rc-tree";


const FolderTree = props => {
    const [open, setOpen] = React.useState(false);
    const [checkedKeys, setCheckedKeys] = React.useState([])

    const onCheck = checkedKeys => {
        setCheckedKeys(checkedKeys)
    }
    const onSelect = selectedKeys => {
        console.log('selected', selectedKeys)
    }
 
    return (
      <Tree
        checkable
        multiple
        onCheck={onCheck}
        onSelect={onSelect}
    >
         <TreeNode title="parent 1" key="0-0"></TreeNode>
    </Tree>

    );
  };
  
export default FolderTree;